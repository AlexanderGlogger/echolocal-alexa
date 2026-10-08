//go:build board_checkers || board_cronos || board_rook

package settings

import (
	sharedpages "github.com/ygelfand/libcountertop/pkg/display/settings"

	"github.com/ygelfand/echolocal/internal/feature/idle"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/visuals"
	"github.com/ygelfand/echolocal/internal/feature/weather"
	"github.com/ygelfand/echolocal/internal/ui/visual"
)

func idlePages() *sharedpages.IdlePages {
	return sharedpages.NewIdlePages(
		sharedpages.IdleDependencies{
			Preferences: preferences(),
			Shell:       shell.Get(),
			Idle:        idle.Get(),
			Media:       media.Get(),
			Visuals:     visuals.Get(),
			Weather:     weather.Get(),
			Thumbnail:   visual.ThumbnailFit,
		},
	)
}

func WeatherPage() *shell.Page { return idlePages().Weather() }
