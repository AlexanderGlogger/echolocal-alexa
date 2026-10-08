//go:build board_checkers || board_cronos || board_rook

package control

import (
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/btcmd"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/a2dp"
)

func init() {
	boardCommands = append(boardCommands, func(t *harness.Tree) {
		t.Top(btcmd.Command(a2dp.Get(), "echolocal"))
	})
	boardSettings = append(boardSettings, speakerRows)
}

func speakerRows() []harness.Setting[config.Config] {
	return []harness.Setting[config.Config]{{
		Name:  "features.speaker",
		Field: "Bluetooth.Speaker",
		Read:  func(c config.Config) string { return harness.OnOff(c.Bluetooth.Speaker) },
		Write: harness.Toggle(a2dp.Get().SetEnabled),
	}}
}
