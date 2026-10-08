package volume

import (
	"log/slog"
	"slices"
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	sharedview "github.com/ygelfand/libcountertop/pkg/display/volume"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(38),
		component.Needs(board.Panel),
	)
}

type Change = sharedvolume.Change

type Volume struct {
	entities *sharedvolume.Entities
	Changed  sharedvolume.Changes

	card    *sharedview.Card
	numbers map[config.Stream]*esphome.Number
}

var (
	once   sync.Once
	shared *Volume
)

func Get() *Volume {
	once.Do(func() {
		shared = &Volume{numbers: map[config.Stream]*esphome.Number{}}
		shared.card = sharedview.NewCard(shell.Get())
		shared.build()

		media.Get().Volume.Listen(func(int) { shared.changed(config.StreamMain) })
		media.Get().SetTurns(shared.turn)
	})
	return shared
}

func (v *Volume) Name() string { return "volume" }

func (v *Volume) Entities() []esphome.Entity { return v.entities.Entities() }

func (v *Volume) Restore(c config.Config) {
	for _, s := range config.Streams() {
		if s == config.StreamMain {
			continue
		}
		v.hold(s, c.Volume.Level(s))
		slog.Info("restored", "what", v.numbers[s].ObjectID, "using", c.Volume.Level(s))
	}
}

func (v *Volume) Level(s config.Stream) int {
	if s == config.StreamMain {
		return media.Get().Step() * 100 / media.VolumeSteps
	}
	return config.Get().Volume.Level(s)
}

func (v *Volume) Set(s config.Stream, level int) {
	level = sharedvolume.Clamp(level)
	if v.Level(s) == level {
		return
	}

	if s == config.StreamMain {
		media.Get().Set((level*media.VolumeSteps + 50) / 100)
		return
	}

	v.hold(s, level)
	if err := config.Set().Volume().Level(s, level); err != nil {
		slog.Error("saving a volume failed", "stream", s, "err", err)
	}
	v.changed(s)
}

func (v *Volume) turn(delta int) bool {
	s, ok := v.card.Selected()
	if !ok {
		s, ok = wanted()
	}
	if !ok || s == config.StreamMain {
		return false
	}

	v.Set(s, v.Level(s)+delta*sharedvolume.Step)
	return true
}

type wants interface{ Wants() (config.Stream, bool) }

func wanted() (config.Stream, bool) {
	v, ok := shell.Get().Top().(wants)
	if !ok {
		return "", false
	}
	return v.Wants()
}

func (v *Volume) hold(s config.Stream, level int) {
	speaker.Get().SetLevel(s, sharedvolume.Clamp(level))
	v.numbers[s].Set(float32(sharedvolume.Clamp(level)))
}

func (v *Volume) changed(s config.Stream) {
	v.Changed.Emit(Change{Stream: s, Level: v.Level(s)})
	if showing(s) {
		shell.Get().Redraw()
		return
	}
	v.card.Show(s)
}

func (v *Volume) build() {
	streams := config.Streams()
	streams = slices.DeleteFunc(
		streams,
		func(s config.Stream) bool { return s == config.StreamMain },
	)
	v.entities = sharedvolume.NewEntities(v, streams, component.DevicePlayback)
	v.numbers = v.entities.Numbers
}

func (v *Volume) Watch(stream config.Stream, changed func(int)) func() {
	return v.Changed.Watch(stream, changed)
}
