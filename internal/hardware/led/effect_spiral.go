package led

import (
	"math"
	"time"
)

func spiral(p Palette) Frame {
	const revolution = 2200 * time.Millisecond

	return func(elapsed time.Duration) []Color {
		phase := float64(elapsed%revolution) / float64(revolution)

		out := make([]Color, Segments)
		for i := range out {
			behind := math.Mod(float64(i)/Segments-phase+1, 1)
			f := 1 - behind
			out[i] = p.Shade(f, f)
		}
		return out
	}
}
