//go:build board_checkers || board_cronos || board_rook

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/lib/wake"
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
)

func assistantPages() *sharedsettings.AssistantPages {
	return sharedsettings.NewAssistantPages(
		sharedsettings.AssistantOptions{
			Slots: wakeword.Slots,
			Read: func(slot int) sharedsettings.AssistantWord {
				w := config.Get().Wake.Slot(slot)
				return sharedsettings.AssistantWord{
					ID:        w.ID,
					Threshold: w.Threshold,
					Tone:      w.Tone,
					Delivery:  w.Delivery,
					Look:      w.Look,
				}
			},
			Models:    wake.Lib().Ours,
			Threshold: func(slot int, v float64) error { return config.Set().Wake(slot).Threshold(v) },
			Tone:      func(slot int, v config.Tone) error { return config.Set().Wake(slot).Tone(v) },
			Delivery:  func(slot int, v config.Delivery) error { return config.Set().Wake(slot).Delivery(v) },
			Requested: func(slot int) { wakeword.Requested.Emit(slot) },
			Chime:     wakeword.Chime,
			Look:      func(slot int) shell.View { return lookPage(slot) },
			Push:      shell.Get().Push,
		},
	)
}
