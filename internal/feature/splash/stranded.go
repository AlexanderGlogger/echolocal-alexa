package splash

import (
	bootview "github.com/ygelfand/libcountertop/pkg/display/boot"
	"github.com/ygelfand/libcountertop/pkg/display/panel"
	sharedui "github.com/ygelfand/libcountertop/pkg/display/ui"

	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/ui"
)

func init() { display.Get().Stranded(drawStranded) }

func drawStranded(p *panel.Panel) error {
	bootview.Stranded(sharedui.Of(p), chosen(), ui.Logo().Draw, layout.Version)
	return nil
}
