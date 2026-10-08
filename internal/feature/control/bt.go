//go:build board_checkers || board_cronos || board_rook

package control

import (
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/btcmd"

	"github.com/ygelfand/echolocal/internal/feature/a2dp"
)

func init() {
	boardCommands = append(boardCommands, func(t *harness.Tree) {
		t.Top(btcmd.Command(a2dp.Get(), "echolocal"))
	})
}
