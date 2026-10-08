//go:build board_checkers || board_cronos || board_rook

package control

import (
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/hardware/touch"
)

func configureTouch(options *harness.InputOptions) {
	options.Deliver = func(c harness.Contact) {
		touch.Get().
			Feed(touch.Contact{ID: c.ID, X: c.X, Y: c.Y, Phase: touch.Phase(c.Phase), At: c.At})
	}
}
