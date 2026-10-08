package speaker

import (
	"math"
	"sync"

	"github.com/ygelfand/echolocal/internal/config"
)

type Resampler interface {
	Run(mono []int16, out []int16) []int16

	Reset()

	Clipped() uint64
}

type linear struct{ prev int16 }

func (l *linear) Reset()        { l.prev = 0 }
func (*linear) Clipped() uint64 { return 0 }

func (l *linear) Run(mono []int16, out []int16) []int16 {
	for _, s := range mono {
		step := (float32(s) - float32(l.prev)) / voicePhases
		for p := range voicePhases {
			v := int16(float32(l.prev) + step*float32(p+1))
			out = append(out, v, v)
		}
		l.prev = s
	}
	return out
}

type hold struct{}

func (hold) Reset()          {}
func (hold) Clipped() uint64 { return 0 }

func (hold) Run(mono []int16, out []int16) []int16 {
	for _, s := range mono {
		for range voicePhases {
			out = append(out, s, s)
		}
	}
	return out
}

const (
	sweepMs   = 1500
	sweepFrom = 300.0
	sweepTo   = 7000.0
)

func VoiceSweep() []int16 {
	frames := VoiceRate * sweepMs / 1000
	out := make([]int16, frames)

	var phase float64
	for i := range out {
		hz := sweepFrom + (sweepTo-sweepFrom)*float64(i)/float64(frames)
		phase += 2 * math.Pi * hz / VoiceRate

		env := math.Min(1, math.Min(float64(i), float64(frames-i))/float64(VoiceRate/50))
		out[i] = int16(0.5 * env * math.MaxInt16 * math.Sin(phase))
	}
	return out
}

var (
	resamplersMu sync.RWMutex
	resamplers   = map[config.Resampling]func() Resampler{
		config.ResampleSinc:   func() Resampler { return &sinc{} },
		config.ResampleLinear: func() Resampler { return &linear{} },
		config.ResampleHold:   func() Resampler { return hold{} },
	}
	resamplerOrder = []config.Resampling{
		config.ResampleSinc,
		config.ResampleLinear,
		config.ResampleHold,
	}
)

func Register(r config.Resampling, make func() Resampler) {
	resamplersMu.Lock()
	defer resamplersMu.Unlock()

	if _, seen := resamplers[r]; !seen {
		resamplerOrder = append(resamplerOrder, r)
	}
	resamplers[r] = make
}

func Resamplings() []config.Resampling {
	resamplersMu.RLock()
	defer resamplersMu.RUnlock()
	return append([]config.Resampling(nil), resamplerOrder...)
}

func NewResampler(r config.Resampling) (Resampler, config.Resampling) {
	resamplersMu.RLock()
	make, ok := resamplers[r]
	resamplersMu.RUnlock()

	if !ok {
		return resamplers[config.ResampleSinc](), config.ResampleSinc
	}
	return make(), r
}
