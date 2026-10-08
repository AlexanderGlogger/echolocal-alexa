package asp

import "github.com/ygelfand/libcountertop/pkg/audio/fft"

// The vendor filter is minimum phase: its largest tap is its first.
type fir struct {
	f     *fft.FFT
	h     []complex64
	work  []complex64
	tail  []float32
	block int
}

func newFIR(taps []float32, block int) *fir {
	n := 1
	for n < block+len(taps)-1 {
		n <<= 1
	}

	f := &fir{
		f:     fft.New(n),
		h:     make([]complex64, n),
		work:  make([]complex64, n),
		tail:  make([]float32, n-block),
		block: block,
	}
	for i, t := range taps {
		f.h[i] = complex(t, 0)
	}
	f.f.Forward(f.h)
	return f
}

func (f *fir) reset() { clear(f.tail) }

func (f *fir) process(x []float32) {
	if len(x) != f.block {
		panic("asp: block is not the size the filter was built for")
	}

	for i, v := range f.tail {
		f.work[i] = complex(v, 0)
	}
	for i, v := range x {
		f.work[len(f.tail)+i] = complex(v, 0)
	}

	keep := len(f.tail)
	if keep <= f.block {
		copy(f.tail, x[f.block-keep:])
	} else {
		copy(f.tail, f.tail[f.block:])
		copy(f.tail[keep-f.block:], x)
	}

	f.f.Forward(f.work)
	for i := range f.work {
		f.work[i] *= f.h[i]
	}
	f.f.Inverse(f.work)

	for i := range x {
		x[i] = real(f.work[keep+i])
	}
}
