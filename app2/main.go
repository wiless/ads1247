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
	// for {
	// 	read()
	// 	time.Sleep(100 * time.Millisecond)
	// }

	reset()
	RDATAC() // CONTINOUS READ MODE
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
		time.Sleep(400 * time.Nanosecond)
		drdybar, e := embd.DigitalRead(_DRDY_GPIO)
		if e == nil {
			if drdybar != 1 {
				// its ready
				break
			}
		} else {
			log.Println("Error Reading GPIO_", _DRDY_GPIO)
		}

	}

}

func configure() {
	//
	// SYS0 = 0x03
	// WREG command = 0x40
	// PARM=bytes to be written -1
	// PGA = 000:1, 100:16 ,111:128 (default 1)
	// DR Data output sampling rate SPS : (default 5 SPS), 0000:5, 1000:1000sps, >1xxx:2000 sps
	// regbyte=[0 PGA DR]
	// regbyte := 00001111 // PGA =1, SPS = 2k sps
	var regbyte uint8 = 0x0F // max gain, max sps = 0111 1111 = 7F
	data := []uint8{0x43, 0x00, regbyte}
	var err error

	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x %x %x ", data[0], data[1], data[2])
	data = []uint8{NOP}
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x ", data[0])

}

func reset() {

	data := []uint8{0x06}
	err := bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x ", data[0])

	data[0] = 0xff
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x ", data[0])
	time.Sleep(100 * time.Microsecond)

	/// start RDATAC

}

// RDATA implements 9.5.3.5 RDATA (0001 001x)
func RDATA() {
	data := []uint8{0x12}
	err := bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	fmt.Printf("\n Received %x ", data[:])

	/// SEND NOPS

	data = []uint8{NOP, NOP, NOP}
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	fmt.Printf("\n Received bytes after RDATA %x ", data[:])

}

// RDATAC Read data continuous mode 0001 010x (14h, 15h)
func RDATAC() {
	data := []uint8{0x12}
	err := bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	fmt.Printf("\n Received %x ", data[:])

	/// SEND NOPS

	for {
		WaitTillDRDY()
		data = []uint8{NOP, NOP, NOP}
		err = bus.TransferAndReceiveData(data[:])
		if err != nil {
			log.Println("Error Reading .. ", err)
		}
		fmt.Printf("\n Received bytes after RDATAC %x ", data[:])
	}

}
