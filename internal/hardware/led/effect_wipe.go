package led

import "time"

func wipe(p Palette) Frame {
	const period = 2400 * time.Millisecond

	return func(elapsed time.Duration) []Color {
		x := 2 * float64(elapsed%period) / float64(period)
		if x > 1 {
			x = 2 - x
		}
		return Arc(x, p.Along(x))
	}
}
