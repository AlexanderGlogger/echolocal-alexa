package speaker

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	sharedtone "github.com/ygelfand/libcountertop/pkg/audio/tone"
	"github.com/ygelfand/libcountertop/pkg/hook"
	"github.com/ygelfand/libcountertop/pkg/runtime/safe"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/lib/alsa"
	"github.com/ygelfand/echolocal/internal/lib/asp"
	"github.com/ygelfand/echolocal/internal/service"
)

// The playback codec accepts one format only: 48 kHz, S16_LE, stereo.
const (
	Rate     = 48000
	Channels = 2
	Bits     = 16

	period  = 1024
	periods = 4

	full = 32768

	// Measured from the vendor HAL doing the same on a route change.
	codecSettle = 1100 * time.Millisecond
)

const (
	VoiceRate     = 16000
	VoiceUpsample = Rate / VoiceRate
)

const (
	Card           = 0
	PlaybackDevice = 23

	AmpSwitch = "Ext_Speaker_Amp_Switch"
)

type Player struct {
	startup atomic.Pointer[component.Progress]
	devMu   sync.Mutex
	pb      *alsa.Playback
	mixer   *alsa.Mixer

	OnOutput hook.Hook[Output]

	volume atomic.Uint32
	step   atomic.Int32

	levels levels

	pathMu sync.Mutex
	out    Output
	seq    sequences

	voiceMu    sync.Mutex
	voice      Resampler
	resampling config.Resampling
	splices    atomic.Uint64

	tuning *asp.Tuning
	chain  *asp.Chain
	on     atomic.Bool
	stale  atomic.Bool
	mono   []float32

	fed bool

	tap    atomic.Pointer[tapTo]
	tapBuf []int16

	underruns atomic.Uint64

	deaf atomic.Uint64

	mu      sync.Mutex
	pending []int16

	written atomic.Uint64

	srcMu  sync.Mutex
	src    Source
	srcBuf []int16
}

type Source interface {
	Render(at uint64, out []int16)
}

func (p *Player) Attach(s Source) {
	p.srcMu.Lock()
	defer p.srcMu.Unlock()
	p.src = s
}

func (p *Player) Written() uint64 { return p.written.Load() }

func New() *Player {
	p := &Player{out: DetectOutput()}
	p.voice, p.resampling = NewResampler(config.ResampleSinc)
	p.SetVolume(VolumeSteps)
	p.on.Store(config.DefaultASP)

	if !component.Board().Has(board.ASP) {
		return p
	}
	t, err := asp.Load(asp.VendorDir)
	if err != nil {
		slog.Error("the driver tuning is not available, playing untuned", "err", err)
		return p
	}
	c, err := t.Chain(period)
	if err != nil {
		slog.Error("the driver tuning will not run, playing untuned", "err", err)
		return p
	}
	p.tuning, p.chain, p.mono = t, c, make([]float32, period)
	return p
}

var (
	once   sync.Once
	shared *Player
)

func init() {
	// The amplifier hisses when nothing drives the DAC, and toggling it pops.
	component.Register(component.Hardware, Get, component.Order(6),
		component.Supervise(service.Restart(time.Second, 30*time.Second)))
}

func Get() *Player {
	once.Do(func() { shared = New() })
	return shared
}

func (p *Player) Name() string { return "speaker" }

// The first mixer write enumerates the card, which takes over a second.
func (p *Player) open() error {
	m, err := alsa.OpenMixer(Card)
	if err != nil {
		return fmt.Errorf("speaker: opening mixer: %w", err)
	}

	pb, err := alsa.OpenPlayback(Card, PlaybackDevice, alsa.Config{
		Channels:   Channels,
		Rate:       Rate,
		Format:     alsa.FormatS16_LE,
		Bits:       Bits,
		PeriodSize: period,
		Periods:    periods,
	})
	if err != nil {
		_ = m.Close()
		return fmt.Errorf("speaker: opening playback: %w", err)
	}

	p.devMu.Lock()
	p.pb, p.mixer = pb, m
	p.devMu.Unlock()
	return nil
}

func (p *Player) device() (*alsa.Playback, *alsa.Mixer) {
	p.devMu.Lock()
	defer p.devMu.Unlock()
	return p.pb, p.mixer
}

func (p *Player) route() {
	p.pathMu.Lock()
	defer p.pathMu.Unlock()
	p.apply(p.seq.path[p.out])
}

func (p *Player) amp(on bool) {
	p.pathMu.Lock()
	defer p.pathMu.Unlock()

	if p.out != OutputSpeaker {
		return
	}
	if on {
		p.apply(p.seq.on)
		return
	}
	p.apply(p.seq.off)
}

func (p *Player) Output() Output {
	p.pathMu.Lock()
	defer p.pathMu.Unlock()
	return p.out
}

func (p *Player) setOutput(out Output) {
	p.pathMu.Lock()
	if p.out == out {
		p.pathMu.Unlock()
		return
	}
	was := p.out
	p.out = out
	p.apply(p.seq.off)
	p.apply(p.seq.leave[was])
	p.apply(p.seq.path[out])
	p.pathMu.Unlock()

	time.Sleep(codecSettle)
	p.amp(true)

	p.SetVolume(int(p.step.Load()))
	slog.Info("output changed", "output", out)
	p.OnOutput.Emit(out)
}

func (p *Player) watchJack(ctx context.Context) {
	t := time.NewTicker(jackPoll)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.setOutput(DetectOutput())
		}
	}
}

func (p *Player) choose() {
	_, mixer := p.device()
	if mixer == nil {
		return
	}
	seq := sequencesFor(func(c string) bool {
		_, err := mixer.Find(c)
		return err == nil
	})
	p.pathMu.Lock()
	p.seq = seq
	p.pathMu.Unlock()
}

func (p *Player) sequences() sequences {
	p.pathMu.Lock()
	defer p.pathMu.Unlock()
	return p.seq
}

func (p *Player) apply(seq []kctl) {
	_, mixer := p.device()
	if mixer == nil {
		return
	}

	for _, c := range seq {
		var err error
		switch {
		case c.value != "":
			err = mixer.SetEnum(c.name, c.value)
		case c.blob != nil:
			err = mixer.SetBytes(c.name, c.blob)
		default:
			err = mixer.SetInt(c.name, uint32(c.level))
		}
		if err != nil {
			slog.Error("mixer write failed", "control", c.name, "err", err)
		}
	}
}

func (p *Player) Run(ctx context.Context) error {
	pb, _ := p.device()
	if pb == nil {
		return errors.New("speaker: the playback device is not held")
	}
	buf := make([]byte, period*Channels*Bits/8)

	// The vendor HAL's route change order: route with the amplifier off, let the codec idle, enable the amplifier, feed.
	out := p.Output()
	p.choose()
	p.apply(p.sequences().init)
	p.route()

	select {
	case <-ctx.Done():
		return nil
	case <-time.After(codecSettle):
	}

	p.amp(true)
	slog.Info("playback path up", "output", out)
	p.OnOutput.Emit(out)
	safe.Go("jack watcher", func() { p.watchJack(ctx) })

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}

		p.fill(buf)
		if err := p.send(ctx, pb, buf); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		p.written.Add(period)
	}
}

func (p *Player) send(ctx context.Context, to io.Writer, buf []byte) error {
	for {
		_, err := to.Write(buf)
		if err == nil {
			return nil
		}
		if err != alsa.ErrUnderrun {
			return err
		}
		if n := p.underruns.Add(1); n == 1 || n%100 == 0 {
			slog.Warn("playback underrun", "times", n)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

type Tap interface{ Offer(mix []int16) }

type tapTo struct{ Tap }

func (p *Player) SetTap(t Tap) {
	if t == nil {
		p.tap.Store(nil)
		return
	}
	p.tap.Store(&tapTo{t})
}

func (p *Player) fill(buf []byte) {
	p.mu.Lock()
	take := min(len(p.pending), period*Channels)
	chunk := p.pending[:take]
	p.pending = p.pending[take:]
	if len(p.pending) == 0 {
		p.pending = nil
	}
	p.mu.Unlock()

	if take > 0 && take < period*Channels {
		p.splices.Add(1)
	}

	rendered := p.render()

	// The internal speaker is one driver wired to the right channel.
	mono := p.Output() == OutputSpeaker

	fed := take > 0 || len(rendered) > 0
	drain := fed || p.fed
	p.fed = fed

	tuned := mono && p.chain != nil && p.on.Load() && drain

	// The vendor's volume is two halves: its EQ bucket's gain before the tuning, the curve's attenuation after.
	makeup := float32(1)
	if tuned {
		makeup = float32(p.tuning.Makeup(float64(p.step.Load()) / VolumeSteps))
	}

	tap := p.tap.Load()
	if tap != nil && cap(p.tapBuf) < period*Channels {
		p.tapBuf = make([]int16, period*Channels)
	}

	gain := p.Volume()
	for i, j := 0, 0; i < period*Channels; i, j = i+Channels, j+1 {
		var l, r int32
		if i < len(chunk) {
			l = int32(chunk[i])
		}
		if i+1 < len(chunk) {
			r = int32(chunk[i+1])
		}
		if i < len(rendered) {
			l += int32(rendered[i])
		}
		if i+1 < len(rendered) {
			r += int32(rendered[i+1])
		}
		if tap != nil {
			p.tapBuf[i], p.tapBuf[i+1] = clamp(
				int32(float32(l)*gain),
			), clamp(
				int32(float32(r)*gain),
			)
		}
		if mono {
			l = (l + r) / 2
			r = l
		}
		if tuned {
			p.mono[j] = float32(clamp(l)) / full * makeup
			continue
		}
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(int16(float32(clamp(l))*gain)))
		binary.LittleEndian.PutUint16(buf[(i+1)*2:], uint16(int16(float32(clamp(r))*gain)))
	}
	if tap != nil {
		tap.Offer(p.tapBuf[:period*Channels])
	}
	if !tuned {
		return
	}

	if p.stale.Swap(false) {
		p.chain.Reset()
	}

	// The tuning's limiter holds a fixed ceiling.
	p.chain.Process(p.mono)
	for i, j := 0, 0; i < period*Channels; i, j = i+Channels, j+1 {
		s := int16(p.mono[j] * full * gain)
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
		binary.LittleEndian.PutUint16(buf[(i+1)*2:], uint16(s))
	}
}

func (p *Player) render() []int16 {
	p.srcMu.Lock()
	src := p.src
	p.srcMu.Unlock()

	if src == nil {
		return nil
	}
	if cap(p.srcBuf) < period*Channels {
		p.srcBuf = make([]int16, period*Channels)
	}
	p.srcBuf = p.srcBuf[:period*Channels]
	clear(p.srcBuf)

	src.Render(p.written.Load(), p.srcBuf)
	attenuate(p.srcBuf, p.Level(config.StreamMedia))
	return p.srcBuf
}

func (p *Player) Play(samples []int16) { p.PlayStream(config.StreamMedia, samples) }

func (p *Player) PlayStream(s config.Stream, samples []int16) {
	p.queue(quieter(samples, p.Level(s)))
}

func (p *Player) queue(samples []int16) {
	if pb, _ := p.device(); pb == nil {
		if n := p.deaf.Add(1); n == 1 || n%100 == 0 {
			slog.Warn("audio dropped, no playback device", "times", n)
		}
		return
	}

	p.mu.Lock()
	p.pending = append(p.pending, samples...)
	p.mu.Unlock()
}

func (p *Player) Take() []int16 {
	p.mu.Lock()
	defer p.mu.Unlock()

	pending := p.pending
	p.pending = nil
	return pending
}

func (p *Player) Adjust(rewrite func([]int16)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	rewrite(p.pending)
}

func (p *Player) PlayVoice(mono []int16) {
	p.voiceMu.Lock()
	out := p.voice.Run(mono, make([]int16, 0, len(mono)*VoiceUpsample*Channels))
	p.voiceMu.Unlock()

	p.PlayStream(config.StreamVoice, out)
}

func (p *Player) Drain() {
	p.mu.Lock()
	p.pending = nil
	p.mu.Unlock()

	p.voiceMu.Lock()
	p.voice.Reset()
	p.voiceMu.Unlock()
}

func (p *Player) SetResampling(r config.Resampling) config.Resampling {
	p.voiceMu.Lock()
	defer p.voiceMu.Unlock()

	p.voice, p.resampling = NewResampler(r)
	return p.resampling
}

func (p *Player) SetASP(on bool) bool {
	on = on && p.chain != nil
	if on && !p.on.Load() {
		p.stale.Store(true)
	}
	p.on.Store(on)
	return on
}

func (p *Player) ASP() bool { return p.on.Load() }

func (p *Player) Resampling() config.Resampling {
	p.voiceMu.Lock()
	defer p.voiceMu.Unlock()
	return p.resampling
}

func (p *Player) Splices() uint64 { return p.splices.Load() }

func (p *Player) Underruns() uint64 { return p.underruns.Load() }

func (p *Player) Clipped() uint64 {
	p.voiceMu.Lock()
	defer p.voiceMu.Unlock()
	return p.voice.Clipped()
}

func (p *Player) Queued() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.pending) / Channels
}

type Note = sharedtone.Note

func (p *Player) Beep(freq float64, ms int, level float64) {
	p.Chime(level, Note{Freq: freq, Ms: ms})
}

func (p *Player) Chime(level float64, notes ...Note) {
	p.ChimeStream(config.StreamFeedback, level, notes...)
}

func (p *Player) ChimeStream(s config.Stream, level float64, notes ...Note) {
	var out []int16
	for _, n := range notes {
		out = append(out, tone(n, level)...)
	}
	p.Overlay(quieter(out, p.Level(s)))
}

func (p *Player) Overlay(samples []int16) {
	if pb, _ := p.device(); pb == nil {
		p.queue(samples)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.pending = mix(p.pending, samples)
}

func mix(into, add []int16) []int16 {
	for i, s := range add {
		if i >= len(into) {
			return append(into, add[i:]...)
		}
		into[i] = clamp(int32(into[i]) + int32(s))
	}
	return into
}

func clamp(v int32) int16 {
	switch {
	case v > math.MaxInt16:
		return math.MaxInt16
	case v < math.MinInt16:
		return math.MinInt16
	}
	return int16(v)
}

func tone(n Note, level float64) []int16 {
	frames := Rate * n.Ms / 1000
	out := make([]int16, frames*Channels)

	for i := range frames {
		if n.Freq == 0 {
			continue
		}
		env := math.Min(1, math.Min(float64(i), float64(frames-i))/float64(Rate/200))
		s := int16(level * env * math.MaxInt16 * math.Sin(2*math.Pi*n.Freq*float64(i)/Rate))
		out[i*Channels] = s
		out[i*Channels+1] = s
	}
	return out
}

func (p *Player) SetVolume(step int) {
	step = max(0, min(step, VolumeSteps))
	p.step.Store(int32(step))
	p.volume.Store(math.Float32bits(gainForStep(p.Output(), step)))
}

func (p *Player) Step() int { return int(p.step.Load()) }

func (p *Player) Volume() float32 { return math.Float32frombits(p.volume.Load()) }

func (p *Player) Close() error {
	p.apply(p.sequences().close)

	p.devMu.Lock()
	pb, mixer := p.pb, p.mixer
	p.pb, p.mixer = nil, nil
	p.devMu.Unlock()

	if mixer != nil {
		_ = mixer.Close()
	}
	if pb == nil {
		return nil
	}
	return pb.Close()
}
