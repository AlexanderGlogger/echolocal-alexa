package led

import (
	"math"
	"time"
)

func flicker(p Palette) Frame {
	return func(elapsed time.Duration) []Color {
		t := elapsed.Seconds()
		whole := 0.75 + 0.25*wander(t, 0)

		out := make([]Color, Segments)
		for i := range out {
			w := wander(t, float64(i)*0.7)
			f := whole * (0.8 + 0.2*w)

			out[i] = scale(p.Along(1-w), math.Min(1, f))
		}
		return out
	}
}
