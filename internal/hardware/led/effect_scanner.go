package led

import (
	"math"
	"time"
)

func scanner(p Palette) Frame {
	const (
		sweep = 900 * time.Millisecond

		width = 2.2
	)

	return func(elapsed time.Duration) []Color {
		x := math.Mod(float64(elapsed)/float64(sweep), 2)
		if x > 1 {
			x = 2 - x
		}
		head := x * (Segments - 1)

		out := make([]Color, Segments)
		for i := range out {
			d := ringDist(i, head)
			if d >= width {
				continue
			}
			f := 1 - d/width
			out[i] = p.Shade(d/width, f*f)
		}
		return out
	}
}
