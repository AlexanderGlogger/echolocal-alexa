package led

import (
	"math"
	"time"
)

func bounce(p Palette) Frame {
	const (
		first = 900 * time.Millisecond
		keeps = 0.55

		spent = 0.04
		rest  = 700 * time.Millisecond
	)

	type arc struct {
		at     time.Duration
		d      time.Duration
		height float64
	}

	var arcs []arc
	var cycle time.Duration
	for h := 1.0; h > spent; h *= keeps {
		d := time.Duration(float64(first) * math.Sqrt(h))
		arcs = append(arcs, arc{at: cycle, d: d, height: h})
		cycle += d
	}
	cycle += rest

	return func(elapsed time.Duration) []Color {
		t := elapsed % cycle

		var at float64
		for _, a := range arcs {
			if t < a.at || t >= a.at+a.d {
				continue
			}
			x := 2*float64(t-a.at)/float64(a.d) - 1
			at = a.height * (1 - x*x) * (Segments - 1)
			break
		}

		out := make([]Color, Segments)
		dot(out, at, p.Along(at/(Segments-1)), 1.5)
		return out
	}
}
