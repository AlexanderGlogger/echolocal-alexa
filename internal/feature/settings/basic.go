//go:build board_checkers || board_cronos || board_rook

package settings

import (
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"

	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/diag"
	"github.com/ygelfand/echolocal/internal/feature/poster"
	"github.com/ygelfand/echolocal/internal/feature/privacy"
	"github.com/ygelfand/echolocal/internal/feature/screen"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/timesync"
	"github.com/ygelfand/echolocal/internal/feature/visuals"
)

func basicPages() *sharedsettings.BasicPages {
	return sharedsettings.NewBasicPages(
		sharedsettings.BasicDependencies{
			Preferences: preferences(),
			Shell:       shell.Get(),
			Screen:      screen.Get(),
			Poster:      poster.Get(),
			Visuals:     visuals.Get(),
			Privacy:     privacy.Get(),
			Access:      diag.Get(),
			Zone:        timesync.Get(),
			Dashboard:   dashboard.Get(),
			Clocks:      clocks,
			Idle:        idlePages(),
			UISize:      uiSize,
			PowerPage:   func() shell.View { return powerPage() },
			VisualPage:  func() shell.View { return visualPage() },
		},
	)
}
