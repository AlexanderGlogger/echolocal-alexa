package led

import "time"

func comet(p Palette) Frame {
	const perSegment = 2 * FrameInterval
	tail := []float64{1, 0.55, 0.3, 0.16, 0.08}

	return func(elapsed time.Duration) []Color {
		head := int(elapsed/perSegment) % Segments
		out := make([]Color, Segments)
		for i, f := range tail {
			c := p.Shade(float64(i)/float64(len(tail)-1), f)
			out[((head-i)%Segments+Segments)%Segments] = c
		}
		return out
	}
}
