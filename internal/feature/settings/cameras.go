//go:build !board_biscuit

package settings

import (
	gogui "github.com/go-gui-org/go-gui/gui"

	sharedshell "github.com/ygelfand/libcountertop/pkg/display/shell"

	"github.com/ygelfand/echolocal/internal/feature/cameras"
)

func camerasPage() *sharedshell.Page { return cameras.Get().Page() }

func init() {
	sections = append(
		sections,
		section{label: "cameras.app", glyph: gogui.IconVideo, page: camerasPage},
	)
}
