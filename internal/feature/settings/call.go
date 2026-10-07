//go:build board_checkers || board_cronos

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/call"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
)

func init() {
	features = append(features, feature{
		key:  "call.settings",
		page: callsPage,
	})
}

func callStreamPage() *shell.Page { return callPages().CallStream() }

func callsPage() *shell.Page { return callPages().Calls() }

func callPages() *sharedsettings.CallPages {
	return sharedsettings.NewCallPages(
		sharedsettings.CallOptions{
			Read:          func() config.Call { return config.Get().Call },
			Push:          shell.Get().Push,
			SetStream:     call.Get().SetStream,
			SetIncoming:   call.Get().SetIncoming,
			SetAutoAnswer: call.Get().SetAutoAnswer,
			SetAutoVideo:  call.Get().SetAutoVideo,
			SetPauseWake:  call.Get().SetPauseWake,
		},
	)
}
