//go:build !board_biscuit

package gui

import (
	"github.com/ygelfand/libcountertop/pkg/display/screens"
	"github.com/ygelfand/libcountertop/pkg/display/widgets"

	"github.com/ygelfand/echolocal/internal/feature/apps"
	"github.com/ygelfand/echolocal/internal/feature/cameras"
)

func init() {
	extensions = append(extensions, func(_ *widgets.UI, o *screens.Options) {
		apps.Get().Add(cameras.Get().App())
		o.Views = append(o.Views, cameras.Get().Screen)
	})
}
