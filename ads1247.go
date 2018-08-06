package ads1247

import (
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	"golang.org/x/exp/io/spi"
)

// Power up;Delay for a minimum of 16 ms to allow power supplies to settle and power-on reset to complete;Enable the device by setting the START pin high;Configure the serial interface of the microcontroller to SPI mode 1 (CPOL = 0, CPHA =1);If the CS pin is not tied low permanently, configure the microcontroller GPIO connected to CS as an output;Configure the microcontroller GPIO connected to the DRDY pin as a falling edge triggered interrupt input;Set CS to the device low;Delay for a minimum of tCSSC;Send the RESET command (06h) to make sure the device is properly reset after power up;Delay for a minimum of 0.6 ms;Send SDATAC command (16h) to prevent the new data from interrupting data or register transactions;Write the respective register configuration with the WREG command (40h, 03h, 01h, 00h, 03h and 42h);As an optional sanity check, read back all configuration registers with the RREG command (four bytes from 20h, 03h);Send the SYNC command (04h) to start the ADC conversion;Delay for a minimum of tSCCS;Clear CS to high (resets the serial interface);Loop{ 	Wait for DRDY to transition low;	Take CS low;	Delay for a minimum of tCSSC;	Send the RDATA command (12h);	Send 24 SCLKs to read out conversion data on DOUT/DRDY;	Delay for a minimum of tSCCS;	Clear CS to high;}Take CS low;Delay for a minimum of tCSSC;Send the SLEEP command (02h) to stop conversions and put the device in power-down mode;

const (
	channel = 0
	speed   = 100000
	bpw     = 8
	delay   = 0
)

// var DRDY_PIN int //GPIO_xx

func init() {
	// initialize SPI dev
	// DRDY_PIN = 0
}

type Sample struct {
	TimeStamp time.Time
	CH        int
	Voltage   float64
	Current   float64
	Value     float64 // unknown adc Value

}

type ADS1247 struct {
	adc        *spi.Device
	_DRDY_GPIO int
	_CS_GPIO   int
	drdyPin    embd.DigitalPin
	onSample   func() Sample
}

func (a *ADS1247) Init(drdy, cs int) error {
	devfs := spi.Devfs{Dev: "/dev/spidev0.0", Mode: spi.Mode1, MaxSpeed: 2000000}

	adc, err := spi.Open(&devfs)
	adc.SetBitOrder(spi.MSBFirst)

	if err != nil {
		log.Panic("Unable to open SPI Device 0.0")
		return err
	} else {
		a.adc = adc
	}
	a.SetCS(cs)
	a.SetDRDY(drdy)
	a.onSample = nil
	return nil
}

func (a *ADS1247) Close() {

	embd.CloseGPIO() // close all gpio

}

func (a *ADS1247) readBack() {

	// SYS0 = 0x03

	// PARM=bytes to be written -1
	// PGA = 000:1, 100:16 ,111:128 (default 1)
	// DR Data output sampling rate SPS : (default 5 SPS), 0000:5, 1000:1000sps, >1xxx:2000 sps
	// regbyte=[0 PGA DR]
	// regbyte := 00001111 // PGA =1, SPS = 2k sps
	//var regbyte uint8 = 0x0F // max gain, max sps = 0111 1111 = 7F
	//data := []uint8{0x43, 0x00, regbyte}
	cmd := []byte{0x23, 0x00}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error reading REG ")
	}
	fmt.Println("\n SYS0 Response : %08b", output)
	cmd = []byte{NOP}
	output = make([]byte, len(cmd))
	err = a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\n SYS0 REGISTER output %08b", output[0])

}

func (a *ADS1247) Initialize() {
	a.Reset()
	a.readBack()    // NOT NEED - Delete after verify
	a.Sdatac()      // Stop continous reading mode..
	a.SetChannel(0) // Set to Default channel
	a.readBack()    //  NOT NEED - Delete after verify
	a.Sync()        //  //  NOT NEED - Delete after verify
	time.Sleep(100 * time.Millisecond)
}

//SetDRDY sets the GPIO pin used to connect to DRDY of ADS1247
func (a *ADS1247) SetDRDY(gpiopin int) {
	// set GPIO mode to input
	a._DRDY_GPIO = gpiopin
	var err error
	a.drdyPin, err = embd.NewDigitalPin(a._DRDY_GPIO)
	if err != nil {
		log.Panic("Unable to Open DRDY Pin", err)
	}

	err = embd.SetDirection(a._DRDY_GPIO, embd.In)
	if err != nil {
		log.Panic("Unable to Set Direction DRDY ", err)
	}

	a.drdyPin.ActiveLow(false)

}

//SetCS sets the GPIO pin used to connect to DRDY of ADS1247
func (a *ADS1247) SetCS(gpiopin int) {
	// set GPIO mode to input
	a._CS_GPIO = gpiopin

	err := embd.SetDirection(a._CS_GPIO, embd.Out)
	if err != nil {
		log.Println("Unable to set CS Pin")
	}
	err = embd.DigitalWrite(a._CS_GPIO, embd.Low)
	if err != nil {
		log.Println("Unable to Enable ADS1247 CS Write Failed")
	}

}

func (a *ADS1247) waitForReady(dev *spi.Device) {
	// Polling method
	for {
		drdybar, e := embd.DigitalRead(a._DRDY_GPIO)
		if e == nil {
			if drdybar != 1 {
				// its ready
				break
			}
		} else {
			log.Println("Error Reading GPIO_", a._DRDY_GPIO)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Read()  implements 9.5.3.5 RDATA (0001 001x)
func (a *ADS1247) Read() int32 {
	// //
	// long A2D = 0x0;
	//   SPI.beginTransaction(SPISettings(_SPIclock1MHZ, MSBFIRST, SPI_MODE1));
	//   digitalWrite(_CSpin, LOW);
	//   A2D |= SPI.transfer(0xFF);
	//   A2D <<= 8;
	//   A2D |= SPI.transfer(0xFF);
	//   A2D <<= 8;
	//   A2D |= SPI.transfer(0xFF);
	//   SPI.transfer(0xFF);
	//   digitalWrite(_CSpin, HIGH);
	//   SPI.endTransaction();
	//   // Convert signs if needed
	//   if (A2D & 0x800000) {
	//     A2D |= ~0xFFFFFF;
	//   }
	//   double inV = 1000.0 * _LSB * A2D;
	//   return inV;
	cmd := []byte{0x12}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Unable to Write Register for REading")
	}

	// a.reset() ??

	cmd = []byte{NOP, NOP, NOP, NOP}
	output = make([]byte, len(cmd))
	err = a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	byteword := []byte{output[0], output[1], output[2], output[3]}
	adclevel := binary.BigEndian.Uint32(byteword)
	// _ = adclevel
	// return uint16(data[1]&0x03)<<8 | uint16(data[2]), nil
	//	return float64(adclevel), nil
	fmt.Printf("\n Received bytes after RDATA  %x %x %x %x", output[0], output[1], output[2], output[3])
	fmt.Printf("\n Received bytes after RDATA  %d", byteword)
	fmt.Printf("Before shifting bytes %d", adclevel)
	fmt.Printf("\nREceived value %d", adclevel>>7)
	result := adclevel >> 7
	return int32(result)

}

const RESET = 0x06
const NOP = 0xff

func (a *ADS1247) Reset() {
	cmd := []byte{0x06}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Unable to Read")
	}
	cmd[0] = NOP

	err = a.adc.Tx(cmd, output)
	time.Sleep(100 * time.Microsecond) /// MAY be dleeted
}

func (a *ADS1247) Sync() {
	cmd := []byte{0x04}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Unable to Read")
	}

	time.Sleep(100 * time.Microsecond)
}

func (a *ADS1247) Sdatac() {
	cmd := []byte{0x16}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Unable to Read")
	}
}

func (a *ADS1247) SetChannel(nCH int) {
	var ch0 byte = 0x02 // ADC channel (default input =  00 000 010 -> AIN0(+ve) & AIN1(-ve))
	var ch1 byte = 0x1a // ADC channel (input =  00 011 010 -> AIN3(+ve) & AIN2(-ve)
	var MUX0 byte
	if nCH == 0 {
		MUX0 = ch0
	} else {
		MUX0 = ch1
	}

	/// 9.5.3.9 WREG (0100 rrrr, 0000 nnnn) (2byte command, START_REG, Nbytes)
	WREG := make([]byte, 2)
	WREG[0] = 0x40        // Start Write Registration from offset 0X00 (0100 rr=0000),
	WREG[1] = 0x03        /// = Write NREGS+1 bytes nnnn=3
	var VBIAS byte = 0x00 // Set to Chip's default = 0000,VBIAS[3:0]= NO BIAS enabled for AIN0:3
	var MUX1 byte = 0x00  // Set to Chip's default = 0,VREFCON[1:0],REFSELT[1:0],MUXCAL[2:0]
	var SYS0 byte = 0x02  // [0, PGA, DR] = 0, 000, 0010
	// PGA = GAIN 000:1, 100:16 ,111:128 (default 1)
	// DR Data output sampling rate SPS : (default 5 SPS), 0000:5, 0010:20SPS, 1000:1000sps, >1xxx:2000 sps
	cmd := []byte{WREG[0], WREG[1], MUX0, VBIAS, MUX1, SYS0}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\nFound this output %x ", output)
	cmd = []byte{NOP}
	output = make([]byte, len(cmd))
	err = a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\nFound this output %x ", output)
	time.Sleep(10 * time.Millisecond)

}

func (a *ADS1247) Configure() {
	//
	// SYS0 = 0x03
	// WREG command = 0x40
	// PARM=bytes to be written -1
	// PGA = 000:1, 100:16 ,111:128 (default 1)
	// DR Data output sampling rate SPS : (default 5 SPS), 0000:5, 1000:1000sps, >1xxx:2000 sps
	// regbyte=[0 PGA DR]
	// regbyte := 00001111 // PGA =1, SPS = 2k sps
	//var regbyte byte = 0x0F // max gain, max sps = 0111 1111 = 7F

	var ch0 byte = 0x02 // ADC channel (default input =  00 000 010 -> AIN0(+ve) & AIN1(-ve))
	var ch1 byte = 0x1a // ADC channel (input =  00 011 010 -> AIN3(+ve) & AIN2(-ve)
	_ = ch1
	cmd := []byte{0x40, 0x03, ch0, 0x00, 0x00, 0x02}
	output := make([]byte, len(cmd))
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\nFound this output %x ", output)
	cmd = []byte{NOP}
	output = make([]byte, len(cmd))
	err = a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\nFound this output %x ", output)
	time.Sleep(10 * time.Millisecond)

}

func (a *ADS1247) ReadSample() Sample {
	a.WaintUntilDRDY()
	var result Sample
	result.TimeStamp = time.Now()
	result.Value = float64(a.Read()) // Actual ADC to Voltage/Current to be done here
	return result
}

// ReadSampleCH reads a ADC sample from the given input channel
func (a *ADS1247) ReadSampleCH(nCH int) Sample {
	a.WaintUntilDRDY()
	var result Sample
	result.TimeStamp = time.Now()
	result.Value = float64(a.Read()) // Actual ADC to Voltage/Current to be done here
	return result
}
