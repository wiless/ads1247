package ads1247

import (
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

// Power up;Delay for a minimum of 16 ms to allow power supplies to settle and power-on reset to complete;Enable the device by setting the START pin high;Configure the serial interface of the microcontroller to SPI mode 1 (CPOL = 0, CPHA =1);If the CS pin is not tied low permanently, configure the microcontroller GPIO connected to CS as an output;Configure the microcontroller GPIO connected to the DRDY pin as a falling edge triggered interrupt input;Set CS to the device low;Delay for a minimum of tCSSC;Send the RESET command (06h) to make sure the device is properly reset after power up;Delay for a minimum of 0.6 ms;Send SDATAC command (16h) to prevent the new data from interrupting data or register transactions;Write the respective register configuration with the WREG command (40h, 03h, 01h, 00h, 03h and 42h);As an optional sanity check, read back all configuration registers with the RREG command (four bytes from 20h, 03h);Send the SYNC command (04h) to start the ADC conversion;Delay for a minimum of tSCCS;Clear CS to high (resets the serial interface);Loop{ 	Wait for DRDY to transition low;	Take CS low;	Delay for a minimum of tCSSC;	Send the RDATA command (12h);	Send 24 SCLKs to read out conversion data on DOUT/DRDY;	Delay for a minimum of tSCCS;	Clear CS to high;}Take CS low;Delay for a minimum of tCSSC;Send the SLEEP command (02h) to stop conversions and put the device in power-down mode;

// Waits for DRDY in blocking mode
func (a *ADS1247) WaintUntilDRDY() {
	done := make(chan bool)

	err := a.drdyPin.Watch(embd.EdgeFalling, func(btn embd.DigitalPin) {
		done <- true
	})
	if err != nil {
		log.Println("Error setting DRDY Watch ", err)
	}
	// This will block the WatchMode till TRIGGERED
	<-done
	err = a.drdyPin.StopWatching()
	if err != nil {
		log.Println("Error setting DRDY Watch ", err)
	}
}

func (a *ADS1247) Notify() chan Sample {
	ch := make(chan Sample) // buffer length
	var s Sample
	err := a.drdyPin.Watch(embd.EdgeFalling, func(btn embd.DigitalPin) {
		a.Read()

		s.TimeStamp = time.Now()
		s.Value = float64(a.Read())
		ch <- s
	})
	if err != nil {
		log.Println("Unable to Add Notifier to DRDY PIN ", err)
	}

	return ch
}
