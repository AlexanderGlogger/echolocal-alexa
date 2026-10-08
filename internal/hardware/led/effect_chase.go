package led

import "time"

func chase(p Palette) Frame {
	const (
		step    = 110 * time.Millisecond
		spacing = 3
	)

	return func(elapsed time.Duration) []Color {
		out := make([]Color, Segments)
		off := int(elapsed/step) % spacing
		for lamp := range Segments / spacing {
			out[off+lamp*spacing] = p.Nth(lamp)
		}
		return out
	}
}
