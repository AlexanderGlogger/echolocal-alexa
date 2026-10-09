//go:build with_screen

package settings

import (
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
	sharedshell "github.com/ygelfand/libcountertop/pkg/display/shell"

	"github.com/ygelfand/echolocal/internal/feature/alerts"
	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/diag"
	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	"github.com/ygelfand/echolocal/internal/feature/homecontrol"
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
			Home:        homecontrol.Get(),
			HA:          homeassistant.Get(),
			Marks:       modeMarks(),
			Alerts:      alerts.Get(),
			Previews:    previewKit(),
			PowerPage:   func() sharedshell.View { return powerPage() },
			VisualPage:  func() sharedshell.View { return visualPage() },
		},
	)
}
