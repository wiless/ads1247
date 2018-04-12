package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/all"
)

var bus embd.SPIBus

var _DRDY_GPIO int = 22 // GPIO_22 (PHYSICAL PIN #15)
const (
	channel = 0
	speed   = 2000000
	bpw     = 8
	delay   = 0
)

func InitSPI() {
	if err := embd.InitSPI(); err != nil {
		log.Println("Unable to Init SPI ", err)
	}
	bus = embd.NewSPIBus(embd.SPIMode0, channel, speed, bpw, delay)

}

var DRDY_PIN, CS_PIN int

func main() {
	InitSPI()
	// sample code to read ADS 1247 analag samples

}

const NOP = 0xff

func read() float64 {

	WaitTillDRDY()

	// ads.reset() ??

	data := [3]uint8{NOP, NOP, NOP}
	/// SIGN CHANGE needs to be verified if output[0] is negative

	// data := [3]uint8{1, 160, 0}

	var err error
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x %x %x ", data[0], data[1], data[2])

	return 0

}

// waits till DRDY pin is not READY "LOW"
func WaitTillDRDY() {

	for {
		drdybar, e := embd.DigitalRead(_DRDY_GPIO)
		if e == nil {
			if drdybar != 1 {
				// its ready
				break
			}
		} else {
			log.Println("Error Reading GPIO_", _DRDY_GPIO)
		}
		time.Sleep(100 * time.Millisecond)
	}

}
