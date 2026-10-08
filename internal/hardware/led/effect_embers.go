package led

import (
	"math"
	"time"
)

func embers(p Palette) Frame {
	base := flicker(p)

	const (
		every = 9 * time.Second
		lasts = 260 * time.Millisecond
	)

	phases := make([]float64, Segments)
	for i := range phases {
		phases[i] = float64(hash(uint32(i)+0x51ED2701)) / float64(1<<32)
	}

	return func(elapsed time.Duration) []Color {
		out := base(elapsed)
		for i, phase := range phases {
			x := math.Mod(float64(elapsed)/float64(every)+phase, 1)
			if x >= float64(lasts)/float64(every) {
				continue
			}

			f := x / (float64(lasts) / float64(every))
			f = math.Sin(math.Pi * math.Pow(f, 0.7))
			out[i] = add(out[i], scale(p.Along(0), 0.7*f))
		}
		return out
	}
}
