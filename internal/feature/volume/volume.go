package volume

import (
	"sync"

	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	sharedview "github.com/ygelfand/libcountertop/pkg/display/volume"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/feedback"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/shell"
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
	Changed sharedvolume.Changes

	card *sharedview.Card
}

var (
	once   sync.Once
	shared *Volume
)

func Get() *Volume {
	once.Do(func() {
		shared = &Volume{card: sharedview.NewCard(shell.Get())}

		media.Get().Levels.Changed.Listen(func(c Change) { shared.changed(c.Stream) })
		media.Get().SetTurns(shared.turn)
	})
	return shared
}

func (v *Volume) Name() string { return "volume" }

func (v *Volume) Level(s config.Stream) int { return media.Get().Levels.Level(s) }

func (v *Volume) Set(s config.Stream, level int) {
	level = sharedvolume.Clamp(level)
	if v.Level(s) == level {
		return
	}
	media.Get().Levels.Set(s, level)
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
	feedback.Preview(s)
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

func (v *Volume) changed(s config.Stream) {
	v.Changed.Emit(Change{Stream: s, Level: v.Level(s)})
	if showing(s) {
		shell.Get().Redraw()
		return
	}
	v.card.Show(s)
}

func (v *Volume) Watch(stream config.Stream, changed func(int)) func() {
	return v.Changed.Watch(stream, changed)
}
