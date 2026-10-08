package led

import (
	"math"
	"time"
)

func ripple(p Palette) Frame {
	const (
		revolution = 3200 * time.Millisecond
		crests     = 2

		glow = 0.12
	)

	return func(elapsed time.Duration) []Color {
		phase := float64(elapsed%revolution) / float64(revolution)

		out := make([]Color, Segments)
		for i := range out {
			c := (1 + math.Cos(2*math.Pi*crests*(float64(i)/Segments-phase))) / 2
			c *= c * c
			out[i] = p.Shade(c, glow+(1-glow)*c)
		}
		return out
	}
}
