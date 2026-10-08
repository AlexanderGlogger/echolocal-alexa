package asp

import (
	"fmt"
	"math"
	"path/filepath"
)

const VendorDir = "/vendor/etc/audio-algorithms"

// The rate the tuning was designed at and the only rate the playback codec takes.
const Rate = 48000

const taps = 1024

// The six vendor EQ files hold one filter shape and differ only in gain.
var eqFiles = []struct {
	upTo float64
	name string
}{
	{0.5, "EQ_50.cfg"},
	{0.6, "EQ_60.cfg"},
	{0.7, "EQ_70.cfg"},
	{0.8, "EQ_80.cfg"},
	{0.9, "EQ_90.cfg"},
	{1.0, "EQ_100.cfg"},
}

const mbclFile = "MBCL.cfg"

type Tuning struct {
	taps  []float32
	gains []float64
	mbcl  mbcl
}

func Load(dir string) (*Tuning, error) {
	t := &Tuning{}
	for _, e := range eqFiles {
		h, err := readFloats(filepath.Join(dir, e.name), taps)
		if err != nil {
			return nil, err
		}
		if t.taps == nil {
			t.taps, t.gains = h, []float64{1}
			continue
		}

		g, err := scaleOf(t.taps, h)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.name, err)
		}
		t.gains = append(t.gains, g)
	}

	m, err := readMBCL(filepath.Join(dir, mbclFile))
	if err != nil {
		return nil, err
	}
	t.mbcl = m
	return t, nil
}

func scaleOf(a, b []float32) (float64, error) {
	var ab, aa float64
	for i := range a {
		ab += float64(a[i]) * float64(b[i])
		aa += float64(a[i]) * float64(a[i])
	}
	g := ab / aa

	var off, tot float64
	for i := range a {
		d := float64(b[i]) - g*float64(a[i])
		off += d * d
		tot += float64(b[i]) * float64(b[i])
	}
	if off/tot > 1e-9 {
		return 0, fmt.Errorf(
			"a different filter, not the same one %.2f dB louder",
			20*math.Log10(g),
		)
	}
	return g, nil
}

func (t *Tuning) Makeup(of float64) float64 {
	for i, e := range eqFiles {
		if of <= e.upTo {
			return t.gains[i]
		}
	}
	return t.gains[len(t.gains)-1]
}

type Chain struct {
	fir  *fir
	comp *mbclState
}

func (t *Tuning) Chain(block int) (*Chain, error) {
	if block <= 0 {
		return nil, fmt.Errorf("asp: a block is %d samples", block)
	}
	if t.mbcl.Bypass {
		return nil, fmt.Errorf("asp: %s asks to be bypassed", mbclFile)
	}

	comp, err := newMBCL(t.mbcl, Rate)
	if err != nil {
		return nil, err
	}
	return &Chain{fir: newFIR(t.taps, block), comp: comp}, nil
}

func (c *Chain) Process(x []float32) {
	c.fir.process(x)
	c.comp.process(x)
}

func (c *Chain) Reset() {
	c.fir.reset()
	c.comp.reset()
}
