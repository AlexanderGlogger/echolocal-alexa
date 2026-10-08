package mic

import "math"

// A 36 mm ring of six microphones and one in the middle: at most about 105 µs, 1.7 samples at 16 kHz, across it.
const (
	// Each microphone after ch0 sits a further 60 degrees clockwise; ch6 is in the middle.
	ringRadius   = 0.036
	firstMic     = 108 * math.Pi / 180
	micSpacing   = 60 * math.Pi / 180
	speedOfSound = 343.0

	Beams = 6

	beamHold = 8
)

const delayLine = 8

type tap struct {
	back int
	frac float32
}

type Beamformer struct {
	taps [Beams][Mics]tap

	hist  [Mics][delayLine]float32
	at    int
	beam  int
	votes int
	next  int

	hold int

	prev [Beams]float32
}

func NewBeamformer() *Beamformer {
	b := &Beamformer{hold: beamHold}

	spread := ringRadius / speedOfSound * float64(Rate)
	for i := range Beams {
		bearing := float64(i) * 2 * math.Pi / Beams
		for m := range Mics {
			d := spread
			if m != CenterMic {
				d = spread * (1 + math.Cos(bearing-(firstMic+float64(m)*micSpacing)))
			}

			back := int(d)
			b.taps[i][m] = tap{back: back, frac: float32(d - float64(back))}
		}
	}
	return b
}

func (b *Beamformer) Mix(mics [][]int16) []int16 {
	if len(mics) < Mics || len(mics[0]) == 0 {
		return nil
	}

	out := make([]int16, len(mics[0]))
	b.scan(mics, out, len(out))
	return out
}

func (b *Beamformer) Look(mics [][]int16) {
	if len(mics) < Mics || len(mics[0]) == 0 {
		return
	}
	b.scan(mics, nil, len(mics[0]))
}

func (b *Beamformer) scan(mics [][]int16, out []int16, n int) {
	if n > len(mics[0]) {
		n = len(mics[0])
	}
	energy := [Beams]float64{}

	for i := range n {
		for m := range Mics {
			b.hist[m][b.at] = float32(mics[m][i])
		}

		for beam := range Beams {
			v := b.sum(beam)

			d := v - b.prev[beam]
			b.prev[beam] = v
			energy[beam] += float64(d) * float64(d)

			if out != nil && beam == b.beam {
				out[i] = clamp(v)
			}
		}
		b.at = (b.at + 1) & (delayLine - 1)
	}

	b.steer(energy)
}

func (b *Beamformer) sum(beam int) float32 {
	var acc float32
	for m := range Mics {
		t := b.taps[beam][m]

		older := b.hist[m][(b.at-t.back)&(delayLine-1)]
		oldest := b.hist[m][(b.at-t.back-1)&(delayLine-1)]
		acc += older + (oldest-older)*t.frac
	}
	return acc / Mics
}

func (b *Beamformer) steer(energy [Beams]float64) {
	best := 0
	for i := range Beams {
		if energy[i] > energy[best] {
			best = i
		}
	}

	if best == b.beam {
		b.votes = 0
		return
	}
	if best != b.next {
		b.next, b.votes = best, 0
	}
	b.votes++
	if b.votes >= b.hold {
		b.beam, b.votes = best, 0
	}
}

func (b *Beamformer) Beam() int { return b.beam }

func clamp(v float32) int16 {
	switch {
	case v > math.MaxInt16:
		return math.MaxInt16
	case v < math.MinInt16:
		return math.MinInt16
	}
	return int16(v)
}
