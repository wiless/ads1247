package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/wiless/ads1247"
)

var DRDY_PIN, CS_PIN int

func init() {
	flag.IntVar(&DRDY_PIN, "DRDY", 22, "GPIO Pin connected to DRDY")
	flag.IntVar(&CS_PIN, "CS", 27, "GPIO Pin connected to DRDY")
	flag.Parse()
}

func main() {
	// sample code to read ADS 1247 analag samples
	var adc ads1247.ADS1247
	e := adc.Init(DRDY_PIN, CS_PIN)
	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}

	sample := adc.Read()
	fmt.Println("Sample value is ", sample)
}
