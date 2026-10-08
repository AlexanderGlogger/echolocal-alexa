package led

import (
	"math"
	"time"
)

func roomFire(p Palette, r Room) Frame {
	const (
		resting = 0.3
		coolest = 0.1
	)

	return func(elapsed time.Duration) []Color {
		x := level(r)
		t := elapsed.Seconds()
		whole := 0.75 + 0.25*wander(t, 0)

		out := make([]Color, Segments)
		for i := range out {
			w := wander(t, float64(i)*0.7)

			heat := (coolest + (1-coolest)*x) * (0.55 + 0.45*w)
			f := x * whole * (0.8 + 0.2*w)
			out[i] = scale(p.Along(1-heat), math.Min(1, f))
		}
		return out
	}
}
