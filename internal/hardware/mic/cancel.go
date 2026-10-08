package mic

import (
	"log/slog"

	"github.com/ygelfand/libcountertop/pkg/audio/aec"
	"github.com/ygelfand/libcountertop/pkg/audio/capture"

	"github.com/ygelfand/echolocal/internal/lib/audio"
)

func newSuppressor(c *aec.Canceller) *aec.Preprocessor {
	var echo *aec.MDF
	if c != nil {
		echo = c.Filter(0)
	}
	p, err := aec.NewSuppressor(Rate, echo)
	if err != nil {
		slog.Error("noise suppression unavailable", "err", err)
		return nil
	}
	return p
}

func (s *Source) suppressBlock(block []int16) {
	if err := s.suppress.Run(block); err != nil {
		slog.Error("noise suppression failed", "err", err)
	}
}

// About -60 dBFS.
const refQuiet = 1e-6 * 32768 * 32768

const refHold = Rate / 4

func newCanceller(taps int) *aec.Canceller {
	c, err := aec.NewCanceller(1, taps, Rate)
	if err != nil {
		slog.Error("echo cancellation unavailable", "err", err)
		return nil
	}
	return c
}

func (s *Source) cancelCenter(mics [][]int16, after func(int, []int16)) []int16 {
	mic := mics[CenterMic]
	n := min(len(mic), len(s.ref))
	if cap(s.cancelled) < n {
		s.cancelled = make([]int16, n)
	}
	out := s.cancelled[:n]

	if err := s.cancel.Process(s.ref[:n], [][]int16{mic[:n]}, [][]int16{out}, after); err != nil {
		slog.Error("echo cancellation failed", "err", err)
		return nil
	}
	return out
}

func (s *Source) sounding(raw []byte, n int) bool {
	if cap(s.ref) < n {
		s.ref = make([]int16, n)
	}
	s.ref = s.ref[:n]
	referenceInto(raw, s.ref)

	switch {
	case playing(s.ref):
		s.hold = refHold
	case s.hold > 0:
		s.hold -= n
	default:
		return false
	}
	return true
}

func playing(ref []int16) bool {
	l := capture.Measure(ref)
	return l.Count > 0 && l.Energy/float64(l.Count) > refQuiet
}

func referenceInto(raw []byte, dst []int16) {
	const frameBytes = Channels * Bits / 8

	frames := min(len(raw)/frameBytes, len(dst))
	for f := range frames {
		o := f*frameBytes + Mics*3
		dst[f] = int16(audio.DecodeS24LE3(raw[o:o+3]) >> 8)
	}
	for f := frames; f < len(dst); f++ {
		dst[f] = 0
	}
}

func (s *Source) Cancelling() bool {
	return s.cancel != nil && s.cancel.Active()
}

func (s *Source) ERLE() float64 {
	if s.cancel == nil {
		return 0
	}
	return s.cancel.ERLE(0)
}

func (s *Source) SetCancelling(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cancelling = on
}

func (s *Source) SetAdapting(on bool) { s.adapt.Store(on) }

func (s *Source) Freeze(on bool) { s.frozen.Store(on) }

func (s *Source) Adapting() bool { return s.adapt.Load() && !s.frozen.Load() }

func (s *Source) Frozen() bool { return s.frozen.Load() }

func (s *Source) ResetCancel() { s.reset.Store(true) }
