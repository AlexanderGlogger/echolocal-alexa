// Package volume is how loud each kind of sound is, and the control for it on the panel.
package volume

import (
	"log/slog"
	"slices"
	"sync"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	esphome "github.com/ygelfand/go-esphome-device"
	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	sharedview "github.com/ygelfand/libcountertop/pkg/display/volume"
	"github.com/ygelfand/libcountertop/pkg/hook"
)

func init() {
	component.Register(component.Device, Get, component.Order(38), component.Needs(board.Panel))
}

// Change is a level that just moved.
type Change struct {
	Stream config.Stream
	Level  int
}

type Volume struct {
	entities *sharedvolume.Entities
	// Changed carries every change, however it was made.
	Changed hook.Hook[Change]

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

// Restore puts the levels back where they were left.
func (v *Volume) Restore(c config.Config) {
	v.entities.RestoreDuck()
	for _, s := range config.Streams() {
		if s == config.StreamMain {
			continue
		}
		v.hold(s, c.Volume.Level(s))
		slog.Info("restored", "what", v.numbers[s].ObjectID, "using", c.Volume.Level(s))
	}
}

// Level is how loud one kind of sound is, as a percentage.
func (v *Volume) Level(s config.Stream) int {
	if s == config.StreamMain {
		return media.Get().Step() * 100 / media.VolumeSteps
	}
	return config.Get().Volume.Level(s)
}

// Set changes a level and remembers it.
func (v *Volume) Set(s config.Stream, level int) {
	level = clamp(level)
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

// Step is how far one press of a volume button moves a level.
const Step = 5

// turn moves whichever level is being shown, and leaves the device's own to the player.
func (v *Volume) turn(delta int) bool {
	s, ok := v.card.Selected()
	if !ok {
		s, ok = wanted()
	}
	if !ok || s == config.StreamMain {
		return false
	}

	v.Set(s, v.Level(s)+delta*Step)
	return true
}

// wants is a view that says which level the buttons move while it is the screen being looked at.
type wants interface{ Wants() (config.Stream, bool) }

func wanted() (config.Stream, bool) {
	v, ok := shell.Get().Top().(wants)
	if !ok {
		return "", false
	}
	return v.Wants()
}

// hold applies a level without writing it back to the file it may have come from.
func (v *Volume) hold(s config.Stream, level int) {
	speaker.Get().SetLevel(s, clamp(level))
	v.numbers[s].Set(float32(clamp(level)))
}

// changed says what the level is now, except over a screen already showing it.
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
	streams = slices.DeleteFunc(streams, func(s config.Stream) bool { return s == config.StreamMain })
	v.entities = sharedvolume.NewEntities(v, streams, nil, component.DevicePlayback)
	v.numbers = v.entities.Numbers
}

func clamp(level int) int {
	switch {
	case level < 0:
		return 0
	case level > 100:
		return 100
	}
	return level
}

// Watch follows one stream's level until the returned function is called.
func (v *Volume) Watch(stream config.Stream, changed func(int)) func() {
	return v.Changed.Listen(func(c Change) {
		if c.Stream == stream {
			changed(c.Level)
		}
	})
}
