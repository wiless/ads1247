package ads1247

import (
	"time"
)

// WaintUntilDRDY for DRDY in blocking mode
func (a *ADS1247) WaintUntilDRDY() {
	/// A dummy sleep /wait
	time.Sleep(100 * time.Millisecond)
}

// Notify creates a go channel to notify when sample is ready
func (a *ADS1247) Notify() chan Sample {
	ch := make(chan Sample) // buffer length
	var s Sample
	go func() {
		for {
			/// A dummy sleep /wait
			time.Sleep(100 * time.Millisecond)
			a.Read()

			s.TimeStamp = time.Now()
			s.Value = float64(a.Read())
			ch <- s
		}
	}()

	return ch
}
