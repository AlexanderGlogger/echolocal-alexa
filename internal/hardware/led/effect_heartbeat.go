package led

import (
	"math"
	"time"
)

func heartbeat(p Palette) Frame {
	const period = 1400 * time.Millisecond

	const decay = 170 * time.Millisecond

	beats := []struct {
		at time.Duration
		f  float64
	}{
		{0, 1},
		{200 * time.Millisecond, 0.6},
	}

	return func(elapsed time.Duration) []Color {
		t := elapsed % period

		var f float64
		for _, b := range beats {
			if t < b.at {
				continue
			}
			f = math.Max(f, b.f*math.Exp(-float64(t-b.at)/float64(decay)))
		}

		return shaded(p, f)
	}
}
