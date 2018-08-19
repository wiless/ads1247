package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/wiless/ads1247"
)

var DRDY_PIN, CS_PIN int
var adc ads1247.ADS1247

func init() {
	flag.IntVar(&DRDY_PIN, "DRDY", 22, "GPIO Pin connected to DRDY")
	flag.IntVar(&CS_PIN, "CS", 27, "GPIO Pin connected to DRDY")
	flag.Parse()
}

var METHOD = 1
var interval int = 1

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Println("Closing all pin and terminating program.")
			adc.Close()
			os.Exit(0)
		}
	}()

	var ch int

	fmt.Printf("\n Channel to be Measured : ")
	fmt.Scanf("%d", &ch)

	fmt.Printf("\n Interval in (s) : ")
	fmt.Scanf("%d", &interval)

	if METHOD == 1 {
		// sample code to read ADS 1247 analag samples
		Method1(ch)
	} else {
		Method2(ch)
	}
}

func Method1(ch int) {

	e := adc.Init(DRDY_PIN, CS_PIN)
	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}
	adc.Initialize()
	adc.SetChannel(ch)
	start := time.Now()
	for {
		adc.WaintUntilDRDY() /// BLOCKED waiting..
		sample := adc.ReadSample()
		fmt.Printf("\n %v  : %v", time.Since(start), sample)
		time.Sleep(time.Duration(interval) * time.Second)
	}

}

func Method2(ch int) {
	// free run mode
	var adc ads1247.ADS1247
	e := adc.Init(DRDY_PIN, CS_PIN)

	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}
	adc.Initialize()
	adc.SetChannel(ch)

	sampleCH := adc.Notify() // continuously sample ADC..
	start := time.Now()
	for {
		s := <-sampleCH
		// fmt.Printf("\n %v  : %v", s.TimeStamp, s.Value)
		fmt.Printf("\n %v  : %v", time.Since(start), s.Value)

		time.Sleep(time.Duration(interval) * time.Second)
	}

}

// TimeStamp time.Time
// 	Voltage   float64
// 	Current   float64
// 	Value     float64 // unknown adc Value
// 	NSamples  int
