//go:build !board_biscuit

package control

import (
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/castcmd"

	"github.com/ygelfand/echolocal/internal/feature/chromecast"
)

func init() {
	boardCommands = append(boardCommands, func(t *harness.Tree) {
		t.Top(castcmd.Command(chromecast.Page))
	})
}
