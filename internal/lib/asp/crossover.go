package asp

import "math"

// Linkwitz-Riley low and high halves sum to an allpass, not to unity.
type crossover struct {
	splits []*split
}

type split struct {
	low, high   *lr4
	compensated []int
	allpass     []*biquad
}

func newCrossover(fc []float64, rate int) *crossover {
	c := &crossover{}
	for i, f := range fc {
		s := &split{low: newLR4(f, rate, false), high: newLR4(f, rate, true)}

		for b := 0; b < i; b++ {
			s.compensated = append(s.compensated, b)
			s.allpass = append(s.allpass, newAllpass(f, rate))
		}
		c.splits = append(c.splits, s)
	}
	return c
}

func (c *crossover) reset() {
	for _, s := range c.splits {
		for _, f := range []*biquad{s.low.a, s.low.b, s.high.a, s.high.b} {
			f.z1, f.z2 = 0, 0
		}
		for _, f := range s.allpass {
			f.z1, f.z2 = 0, 0
		}
	}
}

func (c *crossover) process(x []float32, out [][]float32) {
	rest := out[len(out)-1]
	copy(rest, x)

	for i, s := range c.splits {
		band := out[i]
		for j, v := range rest {
			band[j] = s.low.step(v)
			rest[j] = s.high.step(v)
		}
	}

	for _, s := range c.splits {
		for k, b := range s.compensated {
			ap := s.allpass[k]
			band := out[b]
			for j, v := range band {
				band[j] = ap.step(v)
			}
		}
	}
}

type lr4 struct{ a, b *biquad }

func newLR4(fc float64, rate int, high bool) *lr4 {
	return &lr4{a: newButterworth(fc, rate, high), b: newButterworth(fc, rate, high)}
}

func (l *lr4) step(x float32) float32 { return l.b.step(l.a.step(x)) }

type biquad struct {
	b0, b1, b2, a1, a2 float64
	z1, z2             float64
}

func (f *biquad) step(x float32) float32 {
	in := float64(x)
	out := f.b0*in + f.z1
	f.z1 = f.b1*in - f.a1*out + f.z2
	f.z2 = f.b2*in - f.a2*out
	return float32(out)
}

func newButterworth(fc float64, rate int, high bool) *biquad {
	w := 2 * math.Pi * fc / float64(rate)
	cos, sin := math.Cos(w), math.Sin(w)
	alpha := sin / (2 * math.Sqrt2 / 2)
	a0 := 1 + alpha

	f := &biquad{a1: -2 * cos / a0, a2: (1 - alpha) / a0}
	if high {
		f.b0 = (1 + cos) / 2 / a0
		f.b1 = -(1 + cos) / a0
		f.b2 = f.b0
		return f
	}
	f.b0 = (1 - cos) / 2 / a0
	f.b1 = (1 - cos) / a0
	f.b2 = f.b0
	return f
}

func newAllpass(fc float64, rate int) *biquad {
	w := 2 * math.Pi * fc / float64(rate)
	cos, sin := math.Cos(w), math.Sin(w)
	alpha := sin / (2 * math.Sqrt2 / 2)
	a0 := 1 + alpha

	return &biquad{
		b0: (1 - alpha) / a0,
		b1: -2 * cos / a0,
		b2: 1,
		a1: -2 * cos / a0,
		a2: (1 - alpha) / a0,
	}
}
