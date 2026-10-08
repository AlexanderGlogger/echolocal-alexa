package mic

import (
	"log/slog"
	"math"
	"sync/atomic"

	"github.com/ygelfand/libcountertop/pkg/audio/aec"

	"github.com/ygelfand/echolocal/internal/lib/audio"
)

const cancelTaps = 1024

// Must divide FrameSamples.
const speexFrame = 64

const echoSuppressActive = -40

func newSuppressor(c *canceller) *aec.Preprocessor {
	var echo *aec.MDF
	if c != nil {
		echo = c.filter
	}
	p, err := aec.NewPreprocessor(speexFrame, Rate, echo)
	if err != nil {
		slog.Error("noise suppression unavailable", "err", err)
		return nil
	}
	p.EchoSuppressActive = echoSuppressActive
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

type canceller struct {
	filter *aec.MDF

	// speex averages ERLE over the last half second.
	erle atomic.Int64
	best atomic.Int64

	active atomic.Bool

	refE, micE float64
	blocks     int

	out []int16
}

func newCanceller() *canceller {
	f, err := aec.NewMDF(speexFrame, cancelTaps, Rate)
	if err != nil {
		slog.Error("echo cancellation unavailable", "err", err)
		return nil
	}
	return &canceller{filter: f}
}

func (c *canceller) idle() {
	if c.active.Swap(false) {
		slog.Debug("echo cancellation idle",
			"best_db", float64(c.best.Load())/1000, "last_db", float64(c.erle.Load())/1000,
			"ref_dbfs", meanDBFS(c.refE, c.blocks), "mic_dbfs", meanDBFS(c.micE, c.blocks),
			"seconds", math.Round(float64(c.blocks)*float64(FrameSamples)/Rate*10)/10)
		c.erle.Store(0)
		c.best.Store(0)
		c.refE, c.micE, c.blocks = 0, 0, 0
	}
}

func (c *canceller) apply(ref []int16, mics [][]int16, after func(block []int16)) []int16 {
	if !c.active.Swap(true) {
		slog.Debug("echo cancellation running", "engine", "speex", "taps", cancelTaps)
	}
	mic := mics[CenterMic]
	c.refE += power(ref)
	c.micE += power(mic)
	c.blocks++

	n := min(len(mic), len(ref))
	if cap(c.out) < n {
		c.out = make([]int16, n)
	}
	out := c.out[:n]
	for k := 0; k+speexFrame <= n; k += speexFrame {
		got, err := c.filter.Process(mic[k:k+speexFrame], ref[k:k+speexFrame])
		if err != nil {
			slog.Error("echo cancellation failed", "err", err)
			return nil
		}
		copy(out[k:k+speexFrame], got)
		if after != nil {
			after(out[k : k+speexFrame])
		}
	}
	erle := int64(c.filter.ERLE() * 1000)
	c.erle.Store(erle)
	if erle > c.best.Load() {
		c.best.Store(erle)
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
	return len(ref) > 0 && power(ref) > refQuiet
}

func meanDBFS(energy float64, blocks int) float64 {
	if blocks == 0 || energy <= 0 {
		return -120
	}
	return math.Round(10*math.Log10(energy/float64(blocks)/(32768*32768))*10) / 10
}

func power(s []int16) float64 {
	var sum float64
	for _, v := range s {
		sum += float64(v) * float64(v)
	}
	if len(s) == 0 {
		return 0
	}
	return sum / float64(len(s))
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
	return s.cancel != nil && s.cancel.active.Load()
}

func (s *Source) ERLE() float64 {
	if s.cancel == nil {
		return 0
	}
	return float64(s.cancel.erle.Load()) / 1000
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
