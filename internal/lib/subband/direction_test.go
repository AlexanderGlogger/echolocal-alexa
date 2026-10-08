package subband

import (
	"math"
	"math/rand/v2"
	"testing"
)

// A 36 mm ring of six microphones 60 degrees apart from 108 degrees, and a seventh in the middle.
const (
	ringRadius   = 0.036
	firstMic     = 108 * math.Pi / 180
	micSpacing   = 60 * math.Pi / 180
	speedOfSound = 343.0
	centerMic    = 6

	arrayMics = 7
)

func planeWave(bearing float64, samples int) [][]int16 {
	const tones = 24

	freq := make([]float64, tones)
	phase := make([]float64, tones)
	for i := range freq {
		freq[i] = 300 + 200*float64(i)
		phase[i] = rand.Float64() * 2 * math.Pi
	}

	out := make([][]int16, arrayMics)
	for m := range out {
		out[m] = make([]int16, samples)

		delay := ringRadius / speedOfSound
		if m != centerMic {
			delay *= 1 - math.Cos(bearing-(firstMic+float64(m)*micSpacing))
		}

		for i := range out[m] {
			t := float64(i)/16000 - delay
			var v float64
			for k := range freq {
				v += math.Sin(2*math.Pi*freq[k]*t + phase[k])
			}
			out[m][i] = int16(1000 * v / tones * 8)
		}
	}
	return out
}

func TestVendorBeamsPointSomewhere(t *testing.T) {
	w := vendorWeights(t)

	t.Log("bearing   beam energies (dB relative to the quietest)      loudest")
	pointing := make([]int, 0, 12)

	for step := range 12 {
		bearing := float64(step) * 30 * math.Pi / 180
		frame := planeWave(bearing, 16*testFrame)

		b := w.New()
		b.Mix(frame)

		least := math.MaxFloat64
		for _, e := range b.energy {
			least = min(least, float64(e))
		}

		loudest, line := 0, ""
		for j, e := range b.energy {
			if e > b.energy[loudest] {
				loudest = j
			}
			line += formatDB(float64(e) / least)
		}
		pointing = append(pointing, loudest)
		t.Logf("%4.0f deg  %s   beam %d", float64(step)*30, line, loudest)
	}

	seen := map[int]bool{}
	for _, j := range pointing {
		seen[j] = true
	}
	if len(seen) != w.Beams() {
		t.Errorf("%d of %d beams ever won; the weights are not steering", len(seen), w.Beams())
	}

}

func formatDB(ratio float64) string {
	db := 10 * math.Log10(ratio)
	switch {
	case db >= 10:
		return "  +++"
	case db >= 3:
		return "   ++"
	case db >= 1:
		return "    +"
	}
	return "    ."
}
