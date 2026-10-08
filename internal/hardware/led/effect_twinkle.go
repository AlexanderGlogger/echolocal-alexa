package led

import (
	"math"
	"time"
)

func twinkle(p Palette) Frame {
	const (
		shortest = 1700 * time.Millisecond
		spread   = 1300 * time.Millisecond

		duty = 0.35
		dark = 0.05
	)

	sparks := make([]struct {
		period time.Duration
		phase  float64
		color  Color
	}, Segments)

	for i := range sparks {
		h := hash(uint32(i) + 0x9E3779B9)
		sparks[i].period = shortest + time.Duration(float64(spread)*float64(h&0xFFFF)/0x10000)
		sparks[i].phase = float64(h>>16) / 0x10000
		sparks[i].color = p.At(sparks[i].phase)
	}

	return func(elapsed time.Duration) []Color {
		out := make([]Color, Segments)
		for i, s := range sparks {
			f := dark
			if x := math.Mod(float64(elapsed)/float64(s.period)+s.phase, 1); x < duty {
				f = math.Max(dark, math.Sin(math.Pi*x/duty))
			}
			out[i] = scale(s.color, f)
		}
		return out
	}
}
