package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/wiless/ads1247"
)

var DRDY_PIN, CS_PIN int
var adc ads1247.ADS1247

func init() {
	flag.IntVar(&DRDY_PIN, "DRDY", 22, "GPIO Pin connected to DRDY")
	flag.IntVar(&CS_PIN, "CS", 27, "GPIO Pin connected to DRDY")
	flag.Parse()
}

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

	// sample code to read ADS 1247 analag samples
	Method1()
}

func Method1() {

	e := adc.Init(DRDY_PIN, CS_PIN)
	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}
	adc.Initialize()
	for {
		adc.WaintUntilDRDY() /// BLOCKED waiting..
		sample := adc.ReadSample()
		fmt.Println("Sample value is ", sample)

	}

}

func Method2() {
	// free run mode
	var adc ads1247.ADS1247
	e := adc.Init(DRDY_PIN, CS_PIN)
	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}
	adc.Initialize()
	sampleCH := adc.Notify() // continuously sample ADC..
	for {
		s := <-sampleCH
		fmt.Printf("\n %v  : %v", s.TimeStamp, s.Value)
	}

}

// TimeStamp time.Time
// 	Voltage   float64
// 	Current   float64
// 	Value     float64 // unknown adc Value
// 	NSamples  int
