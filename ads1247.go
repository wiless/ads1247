package ads1247

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	"golang.org/x/exp/io/spi"
	_ github.com/kidoman/embd/host/all
)

// Power up;Delay for a minimum of 16 ms to allow power supplies to settle and power-on reset to complete;Enable the device by setting the START pin high;Configure the serial interface of the microcontroller to SPI mode 1 (CPOL = 0, CPHA =1);If the CS pin is not tied low permanently, configure the microcontroller GPIO connected to CS as an output;Configure the microcontroller GPIO connected to the DRDY pin as a falling edge triggered interrupt input;Set CS to the device low;Delay for a minimum of tCSSC;Send the RESET command (06h) to make sure the device is properly reset after power up;Delay for a minimum of 0.6 ms;Send SDATAC command (16h) to prevent the new data from interrupting data or register transactions;Write the respective register configuration with the WREG command (40h, 03h, 01h, 00h, 03h and 42h);As an optional sanity check, read back all configuration registers with the RREG command (four bytes from 20h, 03h);Send the SYNC command (04h) to start the ADC conversion;Delay for a minimum of tSCCS;Clear CS to high (resets the serial interface);Loop{ 	Wait for DRDY to transition low;	Take CS low;	Delay for a minimum of tCSSC;	Send the RDATA command (12h);	Send 24 SCLKs to read out conversion data on DOUT/DRDY;	Delay for a minimum of tSCCS;	Clear CS to high;}Take CS low;Delay for a minimum of tCSSC;Send the SLEEP command (02h) to stop conversions and put the device in power-down mode;

const (
	channel = 0
	speed   = 2000000
	bpw     = 8
	delay   = 0
)

// var DRDY_PIN int //GPIO_xx

func init() {
	// initialize SPI dev
	// DRDY_PIN = 0
}

type ADS1247 struct {
	adc        *spi.Device
	_DRDY_GPIO int
	_CS_GPIO   int
}

func (ads *ADS1247) Init(drdyPin, csPin int) error {
	devfs := spi.Devfs{Dev: "/dev/spidev.0.0", Mode: spi.Mode1, MaxSpeed: 2000000}

	adc, err := spi.Open(&devfs)
	adc.SetBitOrder(spi.MSBFirst)

	if err != nil {
		log.Panic("Unable to open SPI Device 0.0")
		return err
	} else {
		ads.adc = adc
	}
	ads.SetCS(csPin)
	ads.SetDRDY(drdyPin)
	return nil
}

//SetDRDY sets the GPIO pin used to connect to DRDY of ADS1247
func (ads *ADS1247) SetDRDY(gpiopin int) {
	// set GPIO mode to input
	ads._DRDY_GPIO = gpiopin

	err := embd.SetDirection(ads._DRDY_GPIO, embd.In)
	if err == nil {
		// DRDY_PIN = ads.drdyGPin
	}

}

//SetCS sets the GPIO pin used to connect to DRDY of ADS1247
func (ads *ADS1247) SetCS(gpiopin int) {
	// set GPIO mode to input
	ads._CS_GPIO = gpiopin

	err := embd.SetDirection(ads._CS_GPIO, embd.Out)
	if err != nil {
		log.Println("Unable to set CS Pin")
	}
	err = embd.DigitalWrite(ads._CS_GPIO, embd.Low)
	if err != nil {
		log.Println("Unable to Enable ADS1247 CS Write Failed")
	}

}

func (ads *ADS1247) WaitForReady(dev *spi.Device) {
	for {
		drdybar, e := embd.DigitalRead(ads._DRDY_GPIO)
		if e == nil {
			if drdybar != 1 {
				// its ready
				break
			}
		} else {
			log.Println("Error Reading GPIO_", ads._DRDY_GPIO)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (ads *ADS1247) Read(dev *spi.Device) {
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

	embd.DigitalWrite(ads._CS_GPIO, embd.Low)

	// ads.reset() ??
	var result
	cmd = []byte{NOP,NOP,NOP}
	output=make([]byte,len(cmd)
	err = a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	data:=append([]byte{0},output)
	data := binary.BigEndian.Uint32(mySlice)
    fmt.Println(data) 


}

const RESET = 0x06
const NOP = 0xff

func (a *ADS1247) reset() output {
	cmd := []byte{0x06}
	output:=make([]byte,len(cmd)
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Unable to Read")
	}
	cmd[0] = 0xff

	err = a.adc.Tx(cmd, output)
	time.Sleep(100 * time.Microsecond)
}

func (a *ADS1247) configure() {
	//
	// SYS0 = 0x03
	// WREG command = 0x40
	// PARM=bytes to be written -1
	// PGA = 000:1, 100:16 ,111:128 (default 1)
	// DR Data output sampling rate SPS : (default 5 SPS), 0000:5, 1000:1000sps, >1xxx:2000 sps
	// regbyte=[0 PGA DR]
	// regbyte := 00001111 // PGA =1, SPS = 2k sps
	var regbyte byte = 0x0F // max gain, max sps = 0111 1111 = 7F
	cmd := []byte{0x43, 0x00, regbyte}
	 output:=make([]byte,len(cmd)
	err := a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\nFound this output %x ", output)
	cmd = []byte{NOP}
	output=make([]byte,len(cmd)
	err = a.adc.Tx(cmd, output)
	if err != nil {
		log.Println("Error writing to BUS")
	}
	fmt.Printf("\nFound this output %x ", output)

}

// func (a *ADS1247) reset() {
// }
