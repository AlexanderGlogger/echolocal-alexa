package media

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	esphome "github.com/ygelfand/go-esphome-device"
	"github.com/ygelfand/libcountertop/pkg/audio/noise"
	sharedtone "github.com/ygelfand/libcountertop/pkg/audio/tone"
	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	"github.com/ygelfand/libcountertop/pkg/hook"
	"github.com/ygelfand/libcountertop/pkg/media/ownership"
	"github.com/ygelfand/libcountertop/pkg/runtime/safe"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
	"github.com/ygelfand/echolocal/internal/hardware/led"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func init() {
	component.Register(component.Device, Get, component.Order(25))
}

// Android's STREAM_MUSIC and the vendor volume curves run 0..30.
const VolumeSteps = speaker.VolumeSteps

const volumeFlash = 2 * time.Second

type Player struct {
	mp     *esphome.MediaPlayer
	jack   *esphome.BinarySensor
	stream *Stream

	resampling *esphome.Select
	duck       *esphome.Number

	asp *esphome.Switch

	// The native API has no entity that holds more than one value.
	layers []*esphome.Select

	// Home Assistant's media player state mapper has no case for announcing and raises on it.
	speaking atomic.Bool

	owner     *ownership.Owner[Source]
	ownerInit sync.Once

	Begun hook.Hook[Source]

	Volume hook.Hook[int]

	turns atomic.Pointer[func(delta int) bool]

	step atomic.Int64
}

var (
	once   sync.Once
	shared *Player
)

func Get() *Player {
	once.Do(func() { shared = build() })
	return shared
}

func build() *Player {
	p := &Player{
		mp: &esphome.MediaPlayer{
			Base: esphome.Base{ObjectID: "speaker", Name: "Speaker", Icon: "mdi:speaker"},
			Features: esphome.MediaPlayerFeatureVolumeSet |
				esphome.MediaPlayerFeatureVolumeStep |
				esphome.MediaPlayerFeatureVolumeMute |
				esphome.MediaPlayerFeaturePlayMedia |
				esphome.MediaPlayerFeaturePlay |
				esphome.MediaPlayerFeaturePause |
				esphome.MediaPlayerFeatureStop |
				esphome.MediaPlayerFeatureBrowseMedia |
				esphome.MediaPlayerFeatureAnnounce,
			SupportsPause:    true,
			SupportedFormats: Formats,
		},
		jack: &esphome.BinarySensor{
			Base: esphome.Base{
				ObjectID: "headphones",
				Name:     "Headphones",
				Icon:     "mdi:headphones",
				Category: esphome.CategoryDiagnostic,
			},
			DeviceClass: "plug",
		},
		resampling: &esphome.Select{
			Base: esphome.Base{
				ObjectID: "voice_resampling",
				Name:     "Voice resampling",
				Icon:     "mdi:sine-wave",
				Category: esphome.CategoryConfig,
			},
		},
		duck: sharedvolume.NewDuck(config.MediaSection, component.DevicePlayback),
		asp: &esphome.Switch{
			Base: esphome.Base{
				ObjectID: "speaker_eq",
				Name:     "Speaker EQ",
				Icon:     "mdi:equalizer",
				Category: esphome.CategoryConfig,
			},
		},
	}
	p.layers = noiseLayers()

	bases := []*esphome.Base{
		&p.resampling.Base,
		&p.duck.Base,
		&p.jack.Base,
		&p.asp.Base,
	}
	for _, sel := range p.layers {
		bases = append(bases, &sel.Base)
	}
	for _, b := range bases {
		b.DeviceID = component.DevicePlayback
	}

	p.mp.OnCommand = p.command
	component.Bind(p.resampling, speaker.Resamplings(), speaker.Get().SetResampling,
		config.Set().Speaker().Resampling)

	p.asp.OnCommand = func(want bool) {
		settled := speaker.Get().SetASP(want)
		p.asp.Set(settled)
		if err := config.Set().Speaker().ASP(settled); err != nil {
			slog.Error("saving a setting failed", "setting", p.asp.ObjectID, "err", err)
		}
		slog.Info("setting changed", "setting", p.asp.ObjectID, "using", settled, "asked", want)
	}

	for _, sel := range p.layers {
		sel.OnCommand = func(chosen string) {
			if chosen != noiseOff && !noise.Has(chosen) {
				slog.Warn("unknown option", "setting", sel.ObjectID, "value", chosen)
				return
			}

			sel.Set(chosen)
			p.sound()
		}
	}

	p.stream = NewStream(speaker.Sound(), speaker.Get(), p.refresh)

	buttons.Get().Events.Listen(func(e buttons.Event) {
		if e.Kind == buttons.Hold {
			return
		}
		switch e.Name {
		case buttons.VolumeUp:
			p.turn(1)
		case buttons.VolumeDown:
			p.turn(-1)
		}
	})

	spk := speaker.Get()
	spk.OnOutput.Listen(func(out speaker.Output) { p.jack.Set(out == speaker.OutputHeadphone) })
	p.jack.Set(spk.Output() == speaker.OutputHeadphone)

	p.mp.SetState(esphome.MediaPlayerIdle)
	for _, sel := range p.layers {
		sel.Set(noiseOff)
	}
	return p
}

func (p *Player) Name() string { return "media player" }

func (p *Player) WatchVolume(changed func(int)) func() { return p.Volume.Listen(changed) }

func (p *Player) Entities() []esphome.Entity {
	out := []esphome.Entity{p.mp, p.jack, p.resampling, p.duck, p.asp}
	for _, sel := range p.layers {
		out = append(out, sel)
	}
	return out
}

func (p *Player) Restore(c config.Config) {
	p.apply(c.Speaker.Volume, false)
	slog.Info("restored", "what", "volume", "step", c.Speaker.Volume, "of", VolumeSteps)

	component.Restore(p.resampling, c.Speaker.Resampling, speaker.Get().SetResampling)

	p.duck.Set(float32(c.Media.DuckDB))
	slog.Info("restored", "what", p.duck.ObjectID, "using", c.Media.DuckDB)

	settled := speaker.Get().SetASP(c.Speaker.ASP)
	p.asp.Set(settled)
	slog.Info("restored", "what", p.asp.ObjectID, "using", settled, "asked", c.Speaker.ASP)
}

const noiseOff = "None"

func noiseLayers() []*esphome.Select {
	const slots = 2

	out := make([]*esphome.Select, 0, slots)
	for i := 1; i <= slots; i++ {
		out = append(out, &esphome.Select{
			Base: esphome.Base{
				ObjectID: fmt.Sprintf("noise_layer_%d", i),
				Name:     fmt.Sprintf("White noise layer %d", i),
				Icon:     "mdi:blur",
			},
			Options: append([]string{noiseOff}, noise.Names()...),
		})
	}
	return out
}

func (p *Player) sound() {
	var sounds []string
	for _, sel := range p.layers {
		if chosen := sel.Get(); chosen != noiseOff && chosen != "" {
			sounds = append(sounds, chosen)
		}
	}

	if len(sounds) == 0 {
		p.stream.Stop()
		return
	}
	p.stream.PlayNoise(sounds...)
}

func (p *Player) command(c esphome.MediaCommand) {
	if c.HasVolume {
		p.Set(int(math.Round(float64(c.Volume) * VolumeSteps)))
	}

	if c.HasMediaURL && c.MediaURL != "" {
		if c.Announcement {
			p.announce(c.MediaURL)
		} else {
			p.stream.Play(c.MediaURL)
		}
	}
	if !c.HasCommand {
		return
	}

	switch c.Command {
	case esphome.MediaPlayerVolumeUp:
		p.Adjust(1)
	case esphome.MediaPlayerVolumeDown:
		p.Adjust(-1)
	case esphome.MediaPlayerMute:
		p.Mute(true)
	case esphome.MediaPlayerUnmute:
		p.Mute(false)
	case esphome.MediaPlayerStop:
		p.Stop()
	case esphome.MediaPlayerPause:
		p.Pause()
	case esphome.MediaPlayerPlay:
		p.Unpause()
	case esphome.MediaPlayerToggle:
		if playing, _ := p.Playing(); playing {
			p.Pause()
		} else {
			p.Unpause()
		}
	}
}

func (p *Player) announce(url string) {
	p.Sounding(true)
	claim := speaker.Sound().
		Claim("announce", func(ctx context.Context, spk *speaker.Player) error {
			samples, err := Fetch(ctx, url)
			if err != nil {
				return err
			}
			spk.PlayVoice(samples)
			spk.PlayVoice(make([]int16, speaker.VoiceRate*Tail/1000))
			return nil
		})

	safe.Go("announce", func() {
		<-claim.Done()
		p.Sounding(false)

		if err := claim.Err(); err != nil {
			slog.Error("playing the announcement failed", "url", url, "err", err)
		}
	})
}

func (p *Player) Sounding(on bool) {
	p.speaking.Store(on)
	p.refresh()
}

func (p *Player) External(s Source) {
	if s == nil {
		p.sourceOwner().Clear()
	} else {
		p.sourceOwner().Set(s)
	}
	p.refresh()
}

func (p *Player) Changed() { p.refresh() }

func (p *Player) SetIdle(v config.Delay) {
	if err := config.Set().Idle().Media(v); err != nil {
		slog.Warn("saving the media idle timeout", "err", err)
	}
}

func (p *Player) source() Source { s, _ := p.sourceOwner().Get(); return s }

func (p *Player) Playing() (playing, paused bool) {
	playing, paused = p.stream.Playing()
	if s := p.source(); s != nil {
		now := s.Now()
		ext, extPaused := now.Playing && !now.Paused, now.Paused
		playing, paused = playing || ext, paused || extPaused
	}
	return playing, paused
}

func (p *Player) Pause() {
	p.stream.Pause()
	if s := p.source(); s != nil {
		s.Pause()
	}
}

func (p *Player) Unpause() {
	p.stream.Unpause()
	if s := p.source(); s != nil {
		s.Play()
	}
}

func (p *Player) Stop() {
	p.stream.Stop()
	if s := p.source(); s != nil {
		s.Stop()
	}
}

func (p *Player) refresh() {
	p.mp.SetState(p.state())

	if len(p.stream.Noise()) == 0 {
		for _, sel := range p.layers {
			sel.Set(noiseOff)
		}
	}
}

func (p *Player) state() esphome.MediaPlayerState {
	playing, paused := p.Playing()

	switch {
	case playing || p.speaking.Load():
		return esphome.MediaPlayerPlaying
	case paused:
		return esphome.MediaPlayerPaused
	default:
		return esphome.MediaPlayerIdle
	}
}

func (p *Player) Set(step int) {
	applied := p.apply(step, true)
	if err := config.Set().Speaker().Volume(applied); err != nil {
		slog.Error("saving volume failed", "err", err)
	}
}

func (p *Player) apply(step int, tell bool) int {
	step = max(0, min(step, VolumeSteps))
	p.step.Store(int64(step))

	p.mp.SetVolume(float32(step) / VolumeSteps)
	speaker.Get().SetVolume(step)
	if !tell {
		return step
	}

	p.show(step)
	p.Volume.Emit(step)
	slog.Info("volume", "step", step, "of", VolumeSteps)
	return step
}

func (p *Player) Step() int { return int(p.step.Load()) }

func (p *Player) Mute(muted bool) {
	p.mp.SetMuted(muted)
	if muted {
		speaker.Get().SetVolume(0)
		return
	}
	speaker.Get().SetVolume(p.Step())
}

func (p *Player) SetTurns(fn func(delta int) bool) { p.turns.Store(&fn) }

func (p *Player) turn(delta int) {
	if fn := p.turns.Load(); fn != nil && (*fn)(delta) {
		return
	}
	p.Adjust(delta)
}

func (p *Player) Adjust(delta int) {
	p.Set(p.Step() + delta)
	speaker.Sound().Chime(sharedtone.ToneVolume)
}

func (p *Player) show(step int) {
	leds := led.Get()
	if !leds.Present() {
		return
	}

	frame := led.Volume(float64(step) / VolumeSteps)
	leds.Claim(led.PriorityNotice).PaintFor(frame, volumeFlash)
}

func (p *Player) Start(ctx context.Context) error { p.stream.Start(ctx); return nil }
func (p *Player) Run(ctx context.Context) error   { <-ctx.Done(); return nil }
func (p *Player) Close() error                    { return p.stream.Close() }

func (p *Player) sourceOwner() *ownership.Owner[Source] {
	p.ownerInit.Do(func() {
		p.owner = ownership.New(func(s Source) uint8 { return uint8(kindOf(s)) })
	})
	return p.owner
}
