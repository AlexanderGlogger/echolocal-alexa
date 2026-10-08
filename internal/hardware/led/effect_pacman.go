package led

import (
	"math"
	"time"
)

func pacman(p Palette) Frame {
	const (
		lap = 3600 * time.Millisecond

		chomp = 220 * time.Millisecond
	)

	return func(elapsed time.Duration) []Color {
		at := float64(elapsed%lap) / float64(lap) * Segments

		out := make([]Color, Segments)

		for i := range Segments {
			if i%2 == 0 && float64(i) > at {
				out[i] = p.Nth(1)
			}
		}

		open := math.Abs(math.Sin(math.Pi * float64(elapsed%chomp) / float64(chomp)))
		dot(out, at, scale(p.Nth(0), 0.55+0.45*(1-open)), 1.3)
		return out
	}
}
