package visuals

import (
	sharedview "github.com/ygelfand/libcountertop/pkg/display/visualview"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/gpu"
)

type View = sharedview.View

func (v *Visuals) View() *View {
	return sharedview.New(
		v,
		sharedview.Dependencies{
			GPU:      gpu.Get(),
			Display:  display.Get(),
			Shell:    shell.Get(),
			Settings: config.VisualSection,
		},
	)
}
