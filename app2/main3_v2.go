package main

import (
	"encoding/binary"
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
	speed   = 100000
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
	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	btn, err := embd.NewDigitalPin(_DRDY_GPIO)
	if err != nil {
		panic(err)
	}
	defer btn.Close()

	if err := btn.SetDirection(embd.In); err != nil {
		panic(err)
	}
	btn.ActiveLow(false)

	reset()
	time.Sleep(100 * time.Millisecond)
	readback()
	SDATAC()
	configure()
	time.Sleep(10 * time.Millisecond)
	readback()
	SYNC()
	time.Sleep(100 * time.Millisecond)
	quit := make(chan interface{})
	for i := 0; i < 100; i++ {

		//quit := make(chan interface{})
		err = btn.Watch(embd.EdgeFalling, func(btn embd.DigitalPin) {
			quit <- btn
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("\n%v", <-quit)
		err = btn.StopWatching()
		//		WaitTillDRDY()
		RDATA() //  READ DATA ONCE MODE
		//err=btn.StopWatching()
		if err != nil {
			panic(err)
		}
		// Read every .25 second
		time.Sleep(60 * time.Millisecond)
	}

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
	//var regbyte uint8 = 0x0F // max gain, max sps = 0111 1111 = 7F
	//data := []uint8{0x43, 0x00, regbyte}
	data := []uint8{0x40, 0x03, 0x02, 0x00, 0x30, 0x02}
	//data:=[]uint8{0x43,0x00,0x78}
	var err error

	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x %x %x %x", data[0], data[1], data[2], data[3])

	data = []uint8{NOP}
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	fmt.Printf("\n Received bytes %x ", data[0])

}

func readback() {
	//
	// SYS0 = 0x03
	// WREG command = 0x40
	// PARM=bytes to be written -1
	// PGA = 000:1, 100:16 ,111:128 (default 1)
	// DR Data output sampling rate SPS : (default 5 SPS), 0000:5, 1000:1000sps, >1xxx:2000 sps
	// regbyte=[0 PGA DR]
	// regbyte := 00001111 // PGA =1, SPS = 2k sps
	//var regbyte uint8 = 0x0F // max gain, max sps = 0111 1111 = 7F
	//data := []uint8{0x43, 0x00, regbyte}
	data := []uint8{0x23, 0x00}
	var err error

	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}

	//fmt.Printf("\n Received bytes %x %x %x ", data[0], data[1], data[2])

	//	data = []uint8{NOP, NOP, NOP,NOP}
	data = []uint8{NOP}
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	//fmt.Printf("\n Received bytes after READBACK %x %x %x %x", data[3],data[2],data[1],data[0],)
	fmt.Printf("\n REceived bytes after READBACK %x", data[0])

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

	data = []uint8{NOP, NOP, NOP, NOP}
	err = bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	byteword := []byte{data[0], data[1], data[2], data[3]}
	adclevel := binary.BigEndian.Uint32(byteword)
	// _ = adclevel
	// return uint16(data[1]&0x03)<<8 | uint16(data[2]), nil
	//	return float64(adclevel), nil
	fmt.Printf("\n Received bytes after RDATA  %x %x %x %x", data[0], data[1], data[2], data[3])
	fmt.Printf("\n Received bytes after RDATA  %d", byteword)
	fmt.Printf("Before shifting bytes %d", adclevel)
	fmt.Printf("\nREceived value %d", adclevel>>7)
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

func SDATAC() {
	data := []uint8{0x16}
	err := bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	fmt.Printf("\n Received %x ", data[:])

	/// SEND NOPS
	/*
		for {
			WaitTillDRDY()
			data = []uint8{NOP, NOP, NOP}
			err = bus.TransferAndReceiveData(data[:])
			if err != nil {
				log.Println("Error Reading .. ", err)
			}
			fmt.Printf("\n Received bytes after SDATAC %x ", data[:])
		}*/

}

func SYNC() {
	data := []uint8{0x04}
	err := bus.TransferAndReceiveData(data[:])
	if err != nil {
		log.Println("Error Reading .. ", err)
	}
	fmt.Printf("\n Received %x after SYNC", data[:])

	/// SEND NOPS
	/*
		for {
			WaitTillDRDY()
			data = []uint8{NOP, NOP, NOP}
			err = bus.TransferAndReceiveData(data[:])
			if err != nil {
				log.Println("Error Reading .. ", err)
			}
			fmt.Printf("\n Received bytes after RDATAC %x ", data[:])
		}
	*/
}
