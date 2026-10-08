//go:build board_checkers || board_cronos || board_rook

package control

import (
	"strconv"

	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/screencmd"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/idle"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/network"
	"github.com/ygelfand/echolocal/internal/feature/privacy"
	"github.com/ygelfand/echolocal/internal/feature/screen"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/visuals"
	"github.com/ygelfand/echolocal/internal/feature/weather"
)

var boardSettings []func() []harness.Setting[config.Config]

func settings() []harness.Setting[config.Config] {
	rows := harness.Map(screencmd.Rows(screencmd.Dependencies{
		Clock:     dashboard.Get().Controls,
		Idle:      idle.Get().Idle,
		Visuals:   visuals.Get().Visuals,
		Weather:   weather.Get().Weather,
		Screen:    screen.Get().Screen,
		Settings:  config.ScreenSection,
		Shell:     shell.Get(),
		MediaIdle: media.Get().SetIdle,
		Marks:     privacy.Get().SetMarks,
		Verify:    network.Get().SetVerify,
	}), func(c config.Config) screencmd.State {
		return screencmd.State{
			Network: c.Network,
			Clock:   c.Clock,
			Idle:    c.Idle,
			Screen:  c.Screen,
			Visual:  c.Visual,
			Weather: c.Weather,
		}
	})
	rows = append(rows, harness.Setting[config.Config]{
		Name:  "screen.visual.lift",
		Field: "Microphone.VisualizerLift",
		Read:  func(c config.Config) string { return strconv.Itoa(c.Microphone.VisualizerLift) },
		Write: harness.Number(0, visuals.LiftMax, visuals.Get().SetLift),
	})
	for _, more := range boardSettings {
		rows = append(rows, more()...)
	}
	return rows
}

func setCommand() *cobra.Command {
	return harness.Says(&cobra.Command{
		Use:   "settings [name] [value...]",
		Short: "List the screen's settings, read one, or change one",
	}, set)
}

func set(args []string) (string, error) { return harness.Settings(config.Get, settings(), args) }
