package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/wiless/ads1247"
	"github.com/wiless/gpio"
)

var DRDY_PIN, CS_PIN int
var adc ads1247.ADS1247

func init() {
	flag.IntVar(&DRDY_PIN, "DRDY", 22, "GPIO Pin connected to DRDY")
	flag.IntVar(&CS_PIN, "CS", 6, "GPIO Pin connected to DRDY")
	flag.Parse()
}

var RelayV uint = 17
var RelayC uint = 18
var Vpin, Cpin, CSPin gpio.Pin
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

	fmt.Printf("\n MODE=1 or 2, is Infinite measurement of Voltage (Ch=0) or Current (Ch=1) \n  MODE 3 & 4 uses Voltage & Current Measurement using Relays : ")
	fmt.Scanf("%d", &METHOD)
	if METHOD == 1 || METHOD == 2 {
		fmt.Printf("\n Channel to be Measured : ")
		fmt.Scanf("%d", &ch)

		fmt.Printf("\n Interval in (s) : ")
		fmt.Scanf("%d", &interval)
		log.Println("Starting ... channel : ", ch, " Interval : ", interval, "s")
	} else {
		log.Println("Using GPIO %d for Voltage Relay ", RelayV)
		log.Println("Using GPIO %d for Current Relay ", RelayC)
		Vpin = gpio.NewOutput(RelayV, false)
		Cpin = gpio.NewOutput(RelayC, false)
		CSPin = gpio.NewOutput(uint(CS_PIN), false)

	}

	switch METHOD {
	case 1:
		Method1(ch)
	case 2:
		Method2(ch)
	case 3:
		MeasureVoltageCurrent(0) // type/method  1
	case 4:
		MeasureVoltageCurrent(1) // type/method  2

	}

}

// Measures Voltage and Current with 1 second switch interval
func MeasureVoltageCurrent(method int) {

	if method == 0 {
		Vpin.High()
		Vsamples := MeasureNsamples(0, 10)
		Vpin.Low()
		time.Sleep(20 * time.Millisecond) /// Sleep before switching the next RELAY
		Cpin.High()
		Csamples := MeasureNsamples(1, 100)
		Cpin.Low()
		// time.Sleep(5 * time.Millisecond)
		log.Println(Vsamples)
		log.Println(Csamples)

	}

	if method == 1 {
		Vpin.High()
		Vsamples := MeasureNsamples2(0, 10)
		Vpin.Low()
		Cpin.High()
		Csamples := MeasureNsamples2(1, 100)
		Cpin.Low()

		log.Println(Vsamples)
		log.Println(Csamples)

	}

}

func MeasureNsamples(ch int, Nsamples int) []float64 {
	// var result []float64
	result := make([]float64, Nsamples)
	e := adc.Init(DRDY_PIN, CS_PIN)
	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}
	CSPin.Low() // Note CS pin on Pi is Active Low
	time.Sleep(2 * time.Millisecond)
	adc.Initialize()
	CSPin.High() // Note CS pin on Pi is Active Low
	time.Sleep(2 * time.Millisecond)
	adc.SetChannel(ch)
	start := time.Now()

	for i := 0; i < Nsamples; i++ {

		adc.WaintUntilDRDY() /// BLOCKED waiting..
		CSPin.Low()          // Note CS pin on Pi is Active Low
		time.Sleep(5 * time.Millisecond)

		sample := adc.ReadSample()
		fmt.Printf("\n %v  : %v", time.Since(start), sample)
		time.Sleep(time.Duration(interval) * time.Second)
		CSPin.High() // Note CS pin on Pi is Active Low

		result[i] = sample.Value

	}

	CSPin.Low()

	adc.Sleep()
	time.Sleep(2 * time.Millisecond)
	CSPin.High()
	return result
}

func MeasureNsamples2(ch int, Nsamples int) []float64 {
	// var result []float64
	result := make([]float64, Nsamples)
	var adc ads1247.ADS1247
	e := adc.Init(DRDY_PIN, CS_PIN)

	if e != nil {
		log.Panicln("Unable to Initialize ADC ")
	}
	adc.Initialize()
	adc.SetChannel(ch)

	sampleCH := adc.Notify() // continuously sample ADC..
	start := time.Now()
	for i := 0; i < Nsamples; i++ {

		s := <-sampleCH
		// fmt.Printf("\n %v  : %v", s.TimeStamp, s.Value)
		fmt.Printf("\n %v  : %v", time.Since(start), s.Value)
		time.Sleep(time.Duration(interval) * time.Second)
		result[i] = s.Value

	}
	return result
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
