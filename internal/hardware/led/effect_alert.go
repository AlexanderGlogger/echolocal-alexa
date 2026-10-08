package led

import (
	"math"
	"time"
)

func alert(p Palette) Frame {
	const (
		period = 2 * time.Second
		pulse  = 130 * time.Millisecond
		gap    = 90 * time.Millisecond
		pulses = 3
	)

	return func(elapsed time.Duration) []Color {
		t := elapsed % period

		var f float64
		for i := range pulses {
			at := time.Duration(i) * (pulse + gap)
			if t >= at && t < at+pulse {
				f = math.Sin(math.Pi * float64(t-at) / float64(pulse))
			}
		}
		return shaded(p, f)
	}
}
