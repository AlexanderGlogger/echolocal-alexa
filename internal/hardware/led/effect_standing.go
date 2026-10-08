package led

import (
	"math"
	"time"
)

func standing(p Palette) Frame {
	const (
		period = 2600 * time.Millisecond
		crests = 2
		glow   = 0.08
	)

	return func(elapsed time.Duration) []Color {
		swell := math.Cos(2 * math.Pi * float64(elapsed%period) / float64(period))

		out := make([]Color, Segments)
		for i := range out {
			f := math.Abs(math.Sin(2*math.Pi*crests*float64(i)/Segments) * swell)
			out[i] = p.Shade(f, glow+(1-glow)*f)
		}
		return out
	}
}
