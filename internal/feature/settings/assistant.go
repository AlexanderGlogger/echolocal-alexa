//go:build !board_biscuit

package settings

import (
	"github.com/ygelfand/libcountertop/pkg/assistant/turn"
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/lib/wake"
)

func assistantPages() *sharedsettings.AssistantPages {
	return sharedsettings.NewAssistantPages(
		sharedsettings.AssistantDependencies{
			Slots:       wakeword.Slots,
			Preferences: preferences(),
			Models:      wake.Lib().Ours,
			Requested:   &wakeword.Requested,
			Words:       turn.Words{Settings: config.WakeSection, Speaker: speaker.Sound()},
			Look:        lookPages(),
			Shell:       shell.Get(),
		},
	)
}
