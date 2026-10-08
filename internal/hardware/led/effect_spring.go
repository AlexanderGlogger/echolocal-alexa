package led

import (
	"math"
	"time"
)

func spring(p Palette) Frame {
	const (
		half = 1300 * time.Millisecond

		rings = 1.5
		decay = 3.0

		stretched = 0.72
		slack     = 0.12
	)

	step := func(from, to, x float64) float64 {
		return to + (from-to)*math.Exp(-decay*x)*math.Cos(2*math.Pi*rings*x)
	}

	return func(elapsed time.Duration) []Color {
		t := elapsed % (2 * half)

		var x float64
		if t < half {
			x = step(slack, stretched, float64(t)/float64(half))
		} else {
			x = step(stretched, slack, float64(t-half)/float64(half))
		}

		x = math.Max(0, math.Min(1, x))
		return Arc(x, p.Along(x))
	}
}
