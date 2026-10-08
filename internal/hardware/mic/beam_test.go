//go:build !board_checkers && !board_cronos

package mic

import (
	"math"
	"math/rand/v2"
	"testing"
)

func arrival(m int, bearing float64) float64 {
	if m == CenterMic {
		return 0
	}
	spread := ringRadius / speedOfSound * float64(Rate)
	return -spread * math.Cos(bearing-(firstMic+float64(m)*micSpacing))
}

func simulate(n int, bearing, noise float64, seed uint64) [][]int16 {
	r := rand.New(rand.NewPCG(seed, 7))

	src := make([]float64, n+16)
	for i := range src {
		t := float64(i) / float64(Rate)
		env := 0.6 + 0.4*math.Sin(2*math.Pi*3*t)
		src[i] = env * (3000*math.Sin(2*math.Pi*400*t) +
			2500*math.Sin(2*math.Pi*2200*t) +
			2000*math.Sin(2*math.Pi*3800*t))
	}

	out := make([][]int16, Mics)
	for m := range out {
		out[m] = make([]int16, n)
		d := arrival(m, bearing)
		for i := range n {
			at := float64(i+8) - d
			lo := int(at)
			f := at - float64(lo)
			v := src[lo]*(1-f) + src[lo+1]*f
			out[m][i] = int16(v + noise*(r.Float64()*2-1))
		}
	}
	return out
}

func rms(x []int16) float64 {
	var sum float64
	for _, v := range x {
		sum += float64(v) * float64(v)
	}
	return math.Sqrt(sum / float64(len(x)))
}

func TestBeamformerImprovesSignalToNoise(t *testing.T) {
	const (
		n     = 4800
		noise = 3000
	)

	for _, bearing := range []float64{0, math.Pi / 3, 2 * math.Pi / 3, math.Pi} {
		clean := simulate(n, bearing, 0, 1)
		noisy := simulate(n, bearing, noise, 1)

		b := NewBeamformer()
		b.Mix(simulate(n, bearing, noise, 2))
		got := b.Mix(noisy)

		ref := NewBeamformer()
		ref.Mix(simulate(n, bearing, 0, 2))
		want := ref.Mix(clean)

		var beamErr, micErr float64
		for i := range got {
			d := float64(got[i]) - float64(want[i])
			beamErr += d * d
			d = float64(noisy[CenterMic][i]) - float64(clean[CenterMic][i])
			micErr += d * d
		}
		beamErr, micErr = math.Sqrt(beamErr/float64(n)), math.Sqrt(micErr/float64(n))

		gain := 20 * math.Log10(micErr/beamErr)
		t.Logf(
			"bearing %3.0f°: noise %6.1f on one microphone, %6.1f on the beam, %.1f dB better, beam %d",
			bearing*180/math.Pi,
			micErr,
			beamErr,
			gain,
			b.Beam(),
		)

		if gain < 4 {
			t.Errorf(
				"bearing %.0f°: only %.1f dB better than one microphone",
				bearing*180/math.Pi,
				gain,
			)
		}
		if rms(got) < rms(want)*0.5 {
			t.Errorf("bearing %.0f°: the beam lost the signal, %.0f against %.0f",
				bearing*180/math.Pi, rms(got), rms(want))
		}
	}
}

func TestBeamformerSteersTowardTheSource(t *testing.T) {
	for want := range Beams {
		bearing := float64(want) * 2 * math.Pi / Beams

		b := NewBeamformer()
		for range beamHold + 4 {
			b.Mix(simulate(1600, bearing, 500, 3))
		}
		if got := b.Beam(); got != want {
			t.Errorf("source at %.0f° steered beam %d, want %d", bearing*180/math.Pi, got, want)
		}
	}
}

func TestBeamformerHandlesSilence(t *testing.T) {
	b := NewBeamformer()
	quiet := make([][]int16, Mics)
	for m := range quiet {
		quiet[m] = make([]int16, 320)
	}

	got := b.Mix(quiet)
	for i, v := range got {
		if v != 0 {
			t.Fatalf("sample %d of silence is %d", i, v)
		}
	}
}
