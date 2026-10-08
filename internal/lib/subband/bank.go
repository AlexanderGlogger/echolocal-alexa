package subband

import "github.com/ygelfand/libcountertop/pkg/audio/fft"

type analysis struct {
	g      geometry
	window []float32
	fft    *fft.FFT

	hist []float32
	fold []complex64
}

func newAnalysis(g geometry, window []float32, f *fft.FFT) *analysis {
	return &analysis{
		g:      g,
		window: window,
		fft:    f,
		hist:   make([]float32, g.WindowLen),
		fold:   make([]complex64, g.FFTLen),
	}
}

func (a *analysis) push(samples []float32, out []complex64) {
	copy(a.hist, a.hist[a.g.Hop:])
	copy(a.hist[a.g.WindowLen-a.g.Hop:], samples)

	for i := range a.fold {
		a.fold[i] = 0
	}
	for n, h := range a.window {
		a.fold[n%a.g.FFTLen] += complex(h*a.hist[n], 0)
	}

	a.fft.Forward(a.fold)
	copy(out, a.fold[:a.g.Bands])
}

type synthesis struct {
	g      geometry
	window []float32
	fft    *fft.FFT

	spread []complex64
	tail   []float32
}

func newSynthesis(g geometry, window []float32, f *fft.FFT) *synthesis {
	return &synthesis{
		g:      g,
		window: window,
		fft:    f,
		spread: make([]complex64, g.FFTLen),
		tail:   make([]float32, g.WindowLen),
	}
}

func (s *synthesis) pull(bands []complex64, out []float32) {
	// A real signal's upper half is the lower half conjugate-mirrored; 0 and Nyquist are unpaired.
	s.spread[0] = complex(real(bands[0]), 0)
	for k := 1; k < s.g.Bands; k++ {
		s.spread[k] = bands[k]
		s.spread[s.g.FFTLen-k] = fft.Conj(bands[k])
	}
	s.spread[s.g.Bands] = 0

	s.fft.Inverse(s.spread)

	for n, h := range s.window {
		s.tail[n] += h * real(s.spread[n%s.g.FFTLen])
	}
	copy(out, s.tail[:s.g.Hop])
	copy(s.tail, s.tail[s.g.Hop:])
	for i := s.g.WindowLen - s.g.Hop; i < s.g.WindowLen; i++ {
		s.tail[i] = 0
	}
}

func bankGain(g geometry, window []float32) float32 {
	folded := make([]float32, g.FFTLen)
	for n, h := range window {
		folded[n%g.FFTLen] += h
	}

	var sum float32
	for start := range g.Hop {
		for n := start; n < g.WindowLen; n += g.Hop {
			sum += window[n] * folded[n%g.FFTLen]
		}
	}
	return sum / float32(g.Hop)
}
