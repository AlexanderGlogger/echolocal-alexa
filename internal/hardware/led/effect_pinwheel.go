package led

import (
	"math"
	"time"
)

func pinwheel(p Palette) Frame {
	const (
		revolution = 1600 * time.Millisecond
		arms       = 4

		span = Segments / arms
	)

	return func(elapsed time.Duration) []Color {
		at := float64(elapsed%revolution) / float64(revolution) * Segments

		out := make([]Color, Segments)
		for i := range out {
			turns := (float64(i) - at) / span
			arm := int(math.Round(turns))

			if d := math.Abs(turns-float64(arm)) * span; d < 1 {
				out[i] = scale(p.Nth(arm), 1-d)
			}
		}
		return out
	}
}
