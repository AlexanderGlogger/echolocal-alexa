package splash

import (
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/ui"
	"github.com/ygelfand/echolocal/internal/ui/theme"
	bootview "github.com/ygelfand/libcountertop/pkg/display/boot"
	"github.com/ygelfand/libcountertop/pkg/say"
)

func drawBoot(p *display.Panel, t theme.Theme, progress []component.Progress) {
	bootview.Raster(ui.Of(p), t, progress, layout.Version, say.T("boot.skip"))
}
