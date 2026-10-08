package subband

import (
	"fmt"
	"os"
	"path/filepath"
)

type geometry struct {
	name string

	Bands int

	FFTLen int
	Hop    int

	WindowLen int

	Beams int
	Taps  int

	order    axes
	perGroup int
	split    bool
	tapDown  bool

	mics []int

	// The vendor applies this gain after the beamformer; its weights are not normalised.
	boostDB float64

	windowFile  string
	weightsFile string
}

// The two front ends, from each firmware's own AFE.cfg.
var known = []geometry{
	{
		name:        "FilterBank_640 + FBF",
		Bands:       64,
		FFTLen:      128,
		Hop:         64,
		WindowLen:   640,
		Beams:       6,
		Taps:        4,
		order:       "aetm",
		perGroup:    4,
		split:       true,
		tapDown:     true,
		mics:        []int{0, 1, 2, 3, 4, 5, 6},
		boostDB:     7.2,
		windowFile:  "coefs_FilterBank_640.cfg",
		weightsFile: "coefs_FBF.cfg",
	},
	{
		// AFE.cfg "ASR FixedBeamFormerV2" Channel Map [1,2,4,5]: "This selects 4 mics out of 7".
		name:      "FilterBank_768cvxGLow + FBFV2 8 beams",
		Bands:     128,
		FFTLen:    256,
		Hop:       128,
		WindowLen: 768,
		Beams:     8,
		Taps:      4,
		order:     "maet",
		perGroup:  1,
		// AFE.cfg's "Channel Map" counts channels in Amazon's order, not this array's.
		mics:        []int{0, 3, 2, 1},
		boostDB:     7.2,
		windowFile:  "coefs_FilterBank_AnalysisSynthesis_768cvxGLow.cfg",
		weightsFile: "coefs_FBFV2_LowLatency_8beams.cfg",
	},
}

func (g geometry) Inputs() int { return len(g.mics) }

func (g geometry) Channels() int {
	most := 0
	for _, m := range g.mics {
		most = max(most, m)
	}
	return most + 1
}

func (g geometry) firstSpeechBand() int { return g.Bands / 32 }
func (g geometry) lastSpeechBand() int  { return g.Bands / 2 }

func (g geometry) holdFrames() int { return 160 * 16000 / 1000 / g.Hop }

const VendorDir = "/vendor/etc/audio-algorithms"

type Weights struct {
	g geometry

	window []float32

	fbf []complex64
}

func (w *Weights) weight(band, beam, tap int) []complex64 {
	n := w.g.Inputs()
	at := ((band*w.g.Beams+beam)*w.g.Taps + tap) * n
	return w.fbf[at : at+n]
}

func (w *Weights) Bands() int   { return w.g.Bands }
func (w *Weights) Beams() int   { return w.g.Beams }
func (w *Weights) Name() string { return w.g.name }
func (w *Weights) Inputs() int  { return w.g.Inputs() }

func Load(dir string) (*Weights, error) {
	for _, g := range known {
		if _, err := os.Stat(filepath.Join(dir, g.weightsFile)); err != nil {
			continue
		}
		return g.load(dir)
	}
	return nil, fmt.Errorf("%s holds none of the coefficient sets we know", dir)
}

func (g geometry) load(dir string) (*Weights, error) {
	window, err := readFloats(filepath.Join(dir, g.windowFile), g.WindowLen)
	if err != nil {
		return nil, err
	}
	w := &Weights{g: g, window: window}
	if err := w.readFBF(filepath.Join(dir, g.weightsFile)); err != nil {
		return nil, err
	}
	return w, nil
}
