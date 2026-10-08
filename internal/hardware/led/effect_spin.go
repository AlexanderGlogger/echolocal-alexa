package led

import "time"

func spin(p Palette) Frame {
	const revolution = 1500 * time.Millisecond

	return func(elapsed time.Duration) []Color {
		phase := float64(elapsed%revolution) / float64(revolution)
		out := make([]Color, Segments)
		for i := range out {
			out[i] = p.At(float64(i)/Segments + phase)
		}
		return out
	}
}
