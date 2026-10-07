//go:build board_checkers || board_cronos

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/chromecast"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	sharedcast "github.com/ygelfand/libcountertop/pkg/display/settings/castpages"
)

func servicePages() *sharedcast.Pages {
	return sharedcast.New(sharedcast.Dependencies{Settings: config.CastSection, Shell: shell.Get(), Receiver: chromecast.Get()})
}
