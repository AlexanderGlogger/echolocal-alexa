package reveal

import (
	_ "embed"

	bootview "github.com/ygelfand/libcountertop/pkg/display/boot"
	"github.com/ygelfand/libcountertop/pkg/display/theme"

	"github.com/ygelfand/echolocal/internal/ui"
)

//go:embed reveal.glsl
var revealShader string

func New(palette theme.Theme) *bootview.Reveal {
	return bootview.NewReveal(palette, revealShader, ui.Logo().On(palette.Background), 768, 512)
}
