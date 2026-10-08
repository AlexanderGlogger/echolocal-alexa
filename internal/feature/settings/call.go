//go:build board_checkers || board_cronos

package settings

import (
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
	sharedshell "github.com/ygelfand/libcountertop/pkg/display/shell"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/call"
	"github.com/ygelfand/echolocal/internal/feature/shell"
)

func init() {
	features = append(features, feature{
		key:  "call.settings",
		page: callsPage,
	})
}

func callsPage() *sharedshell.Page { return callPages().Calls() }

func callPages() *sharedsettings.CallPages {
	return sharedsettings.NewCallPages(
		sharedsettings.CallDependencies{
			Settings:   config.CallSection,
			Shell:      shell.Get(),
			Controller: call.Get(),
		},
	)
}
