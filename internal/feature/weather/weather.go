//go:build board_checkers || board_cronos || board_rook

package weather

import (
	"sync"

	"github.com/ygelfand/libcountertop/pkg/homeassistant/weather"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/states"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(61),
		component.Needs(board.Panel),
	)
}

type Weather struct{ *weather.Weather }

func (w Weather) Restore(cfg config.Config) { w.Publish(cfg.Weather) }

var get = sync.OnceValue(func() *Weather {
	return &Weather{weather.New(weather.Options{
		Controller: homeassistant.Get(),
		States:     states.Get(),
		Settings:   config.WeatherSection,
		Home:       config.HomeSection,
		Redraw:     func() { shell.Get().Redraw() },
		DeviceID:   component.DeviceScreen,
	})}
})

func Get() *Weather { return get() }
