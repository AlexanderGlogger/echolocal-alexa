//go:build board_checkers || board_cronos

package settings

import (
	sharedcast "github.com/ygelfand/libcountertop/pkg/display/settings/castpages"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/chromecast"
	"github.com/ygelfand/echolocal/internal/feature/shell"
)

func servicePages() *sharedcast.Pages {
	return sharedcast.New(
		sharedcast.Dependencies{
			Settings: config.CastSection,
			Shell:    shell.Get(),
			Receiver: chromecast.Get(),
		},
	)
}
