package dashboard

import (
	"sync"
	"time"

	sharedlib "github.com/ygelfand/libcountertop/pkg/display/dashboard"
	"github.com/ygelfand/libcountertop/pkg/display/idleview"
	sharedshell "github.com/ygelfand/libcountertop/pkg/display/shell"
	"github.com/ygelfand/libcountertop/pkg/display/theme"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/poster"
	"github.com/ygelfand/echolocal/internal/feature/screen"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/timesync"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(30),
		component.Needs(board.Panel),
	)
}

type Dashboard struct{ *sharedlib.Controls }

var (
	once   sync.Once
	shared *Dashboard
)

func Get() *Dashboard {
	once.Do(func() {
		shared = &Dashboard{sharedlib.NewControls(sharedlib.Dependencies{
			Preferences: sharedlib.Preferences{
				Screen: config.ScreenSection,
				Clock:  config.ClockSection,
			},
			Display:  shell.Get(),
			DeviceID: component.DeviceScreen,
		})}

		component.Settings.Add(shared.Settings())
		screen.Get().Themed.Listen(func(theme.Theme) { shared.Redraw() })
		screen.Get().Restyled.Listen(func(schema.Screen) { shared.Redraw() })
		poster.Get().Changed.Listen(func(int) { shared.Redraw() })
		timesync.Get().WatchSteps(func(time.Duration) { shared.Redraw() })
		shell.Get().Changed.Listen(func(c sharedshell.Change) {
			if _, ok := c.To.(*idleview.View); ok {
				strip.Clock()
			}
		})
	})
	return shared
}

func (d *Dashboard) Name() string { return "dashboard" }

func (d *Dashboard) Restore(cfg config.Config) {
	d.Controls.Restore(cfg.Screen, cfg.Clock)
}

// The device boots with its clock at 1970.
func (d *Dashboard) Ready() bool { return timesync.Get().Valid() }

func (d *Dashboard) Redraw() { shell.Get().Redraw() }
