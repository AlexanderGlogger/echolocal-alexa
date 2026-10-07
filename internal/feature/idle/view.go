package idle

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dashboard/face"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/visuals"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/gpu"
	"github.com/ygelfand/echolocal/internal/ui/visual"
	sharedview "github.com/ygelfand/libcountertop/pkg/display/idleview"
	"time"
)

type View = sharedview.View
type Piece = sharedview.Piece

var Areas = sharedview.Areas
var Pieces = sharedview.Pieces

func newView() *View {
	return sharedview.New(sharedview.Options{
		Read: func() sharedview.Settings {
			c := config.Get()
			return sharedview.Settings{
				Screen: c.Screen,
				Clock:  c.Clock,
				Idle:   c.Idle,
				Visual: c.Visual,
			}
		},
		Hold:        func() func() { return visuals.Get().Hold() },
		Input:       func() visual.Input { return visuals.Get().Input() },
		From:        visuals.From,
		Open:        gpu.Open,
		Orientation: func() int { return int(display.Get().Orientation()) },
		Visible:     shell.Get().Visible,
		Redraw:      shell.Get().Redraw,
	})
}
func Reading(c config.Config, at time.Time) face.Reading {
	return sharedview.Reading(
		sharedview.Settings{Screen: c.Screen, Clock: c.Clock, Idle: c.Idle, Visual: c.Visual},
		at,
	)
}
