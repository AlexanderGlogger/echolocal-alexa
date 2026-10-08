package led

import "time"

func helix(p Palette) Frame {
	const revolution = 2400 * time.Millisecond

	return func(elapsed time.Duration) []Color {
		phase := float64(elapsed%revolution) / float64(revolution) * Segments

		out := make([]Color, Segments)
		dot(out, phase, p.Nth(0), 1.6)
		dot(out, -phase, p.Nth(1), 1.6)
		return out
	}
}
