//go:build board_checkers || board_cronos

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/call"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/lib/say"
	"github.com/ygelfand/echolocal/internal/ui/widget"
)

func init() {
	features = append(features, feature{
		label: func() string { return say.T("call.settings") },
		hint:  func() string { return say.T("call.settings.hint") },
		page:  callsPage,
	})
}

func callStreamPage() *shell.Page {
	return &shell.Page{
		Title: say.T("call.stream"),
		Build: func() ([]widget.Row, []func(int)) {
			now := config.Get().Call.Stream
			var rows []widget.Row
			var acts []func(int)
			for _, v := range config.CallStreams() {
				rows = append(
					rows,
					widget.Row{Label: v.Label(), Kind: widget.Plain, Chosen: v == now},
				)
				acts = append(acts, func(int) { call.SetStream(v) })
			}
			return rows, acts
		},
	}
}

func callsPage() *shell.Page {
	return &shell.Page{
		Title: say.T("call.settings"),
		Build: func() ([]widget.Row, []func(int)) {
			c := config.Get().Call
			rows := []widget.Row{
				{Label: say.T("call.incoming.allow"), Kind: widget.Toggle, On: c.Incoming},
				{
					Label: say.T("call.auto"),
					Hint:  say.T("call.auto.hint"),
					Kind:  widget.Toggle,
					On:    c.AutoAnswer,
				},
			}
			acts := []func(int){
				func(int) { call.SetIncoming(!c.Incoming) },
				func(int) { call.SetAutoAnswer(!c.AutoAnswer) },
			}
			if c.AutoAnswer {
				rows = append(
					rows,
					widget.Row{
						Label: say.T("call.autovideo"),
						Kind:  widget.Toggle,
						On:    c.AutoVideo,
					},
				)
				acts = append(acts, func(int) { call.SetAutoVideo(!c.AutoVideo) })
			}
			rows = append(
				rows,
				widget.Row{Label: say.T("call.pausewake"), Kind: widget.Toggle, On: c.PauseWake},
			)
			acts = append(acts, func(int) { call.SetPauseWake(!c.PauseWake) })
			rows = append(
				rows,
				widget.Row{
					Label: say.T("call.stream"),
					Hint:  say.T("call.stream.hint"),
					Kind:  widget.Chevron,
					Value: c.Stream.Label(),
				},
			)
			acts = append(acts, open(callStreamPage()))
			return rows, acts
		},
	}
}
