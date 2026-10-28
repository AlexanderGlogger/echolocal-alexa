//go:build !board_biscuit

package frontend

import (
	"image"
	"net/http"

	"github.com/ygelfand/libcountertop/pkg/display/desktop"
	"github.com/ygelfand/libcountertop/pkg/input/textentry"
	"github.com/ygelfand/libcountertop/pkg/input/touch"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/screencmd"

	"github.com/ygelfand/echolocal/internal/hardware/display"
	hwtouch "github.com/ygelfand/echolocal/internal/hardware/touch"
)

var shot = func() (image.Image, error) { return screencmd.Composed(display.Get()) }

func addDesktop(pages map[string]http.Handler, input *textentry.Controller) {
	pages["desktop"] = desktop.New(
		display.Get(),
		func(c touch.Contact) { hwtouch.Get().Feed(c) },
		input,
	)
}
