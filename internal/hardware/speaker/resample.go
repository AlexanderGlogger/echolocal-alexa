package speaker

import "math"

const (
	voicePhases = VoiceUpsample
	voiceTaps   = 12
	voiceLength = voicePhases * voiceTaps
)

var voiceFilter = func() [voiceLength]float32 {
	var h [voiceLength]float32
	center := float64(voiceLength-1) / 2
	for i := range h {
		x := (float64(i) - center) / voicePhases
		s := 1.0
		if x != 0 {
			s = math.Sin(math.Pi*x) / (math.Pi * x)
		}
		w := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(voiceLength-1))
		h[i] = float32(s * w * 0.9)
	}
	return h
}()

type sinc struct {
	hist [voiceTaps]float32

	clipped uint64
}

func (u *sinc) Reset()          { u.hist = [voiceTaps]float32{} }
func (u *sinc) Clipped() uint64 { return u.clipped }

func (u *sinc) Run(mono []int16, out []int16) []int16 {
	for _, s := range mono {
		copy(u.hist[:], u.hist[1:])
		u.hist[voiceTaps-1] = float32(s)

		for p := range voicePhases {
			var acc float32
			for k := range voiceTaps {
				acc += voiceFilter[p+k*voicePhases] * u.hist[voiceTaps-1-k]
			}
			if acc > math.MaxInt16 || acc < math.MinInt16 {
				u.clipped++
			}
			v := clamp16(acc)
			out = append(out, v, v)
		}
	}
	return out
}

func clamp16(v float32) int16 {
	switch {
	case v > math.MaxInt16:
		return math.MaxInt16
	case v < math.MinInt16:
		return math.MinInt16
	}
	return int16(v)
}
