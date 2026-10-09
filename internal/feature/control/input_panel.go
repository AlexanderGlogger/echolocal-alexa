//go:build with_screen

package control

import (
	"github.com/ygelfand/libcountertop/pkg/input/touch"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	hwtouch "github.com/ygelfand/echolocal/internal/hardware/touch"
)

func configureTouch(options *harness.InputOptions) {
	options.Deliver = func(c harness.Contact) {
		hwtouch.Get().
			Feed(touch.Contact{ID: c.ID, X: c.X, Y: c.Y, Phase: touch.Phase(c.Phase), At: c.At})
	}
}
