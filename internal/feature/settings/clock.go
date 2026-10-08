//go:build board_checkers || board_cronos || board_rook

package settings

import (
	sharedpages "github.com/ygelfand/libcountertop/pkg/display/settings"

	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/shell"
)

var clocks = sharedpages.NewClockPages(
	sharedpages.ClockDependencies{
		Preferences: preferences(),
		Shell:       shell.Get(),
		Controller:  dashboard.Get(),
	},
)
