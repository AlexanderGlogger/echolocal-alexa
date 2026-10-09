//go:build board_checkers || board_cronos || board_rook

package settings

import (
	sharedpages "github.com/ygelfand/libcountertop/pkg/display/settings"
	sharedshell "github.com/ygelfand/libcountertop/pkg/display/shell"

	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	"github.com/ygelfand/echolocal/internal/feature/idle"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/visuals"
	"github.com/ygelfand/echolocal/internal/feature/weather"
	"github.com/ygelfand/echolocal/internal/ui/visual"
)

func idlePages() *sharedpages.IdlePages {
	return sharedpages.NewIdlePages(
		sharedpages.IdleDependencies{
			Preferences:  preferences(),
			Home:         homeassistant.Get(),
			Marks:        modeMarks(),
			Shell:        shell.Get(),
			Idle:         idle.Get(),
			Visuals:      visuals.Get(),
			Weather:      weather.Get(),
			SaverWeather: weather.Get().Saver(),
			Thumbnail:    visual.Thumbs().Fit,
		},
	)
}

func WeatherEntities() *sharedshell.Page { return idlePages().WeatherEntities() }
