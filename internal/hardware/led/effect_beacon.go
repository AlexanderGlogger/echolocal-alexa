package led

import "time"

func beacon(p Palette) Frame {
	const (
		revolution = 900 * time.Millisecond
		width      = 2.5
	)

	return func(elapsed time.Duration) []Color {
		at := float64(elapsed%revolution) / float64(revolution) * Segments

		out := make([]Color, Segments)
		for i := range out {
			d := ringDist(i, at)
			if d >= width {
				continue
			}
			f := 1 - d/width
			out[i] = p.Shade(f, f*f)
		}
		return out
	}
}
