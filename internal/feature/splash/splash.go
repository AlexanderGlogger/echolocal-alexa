package splash

import (
	"context"
	"sync"
	"sync/atomic"

	bootview "github.com/ygelfand/libcountertop/pkg/display/boot"
	"github.com/ygelfand/libcountertop/pkg/display/panel"
	"github.com/ygelfand/libcountertop/pkg/display/style"
	"github.com/ygelfand/libcountertop/pkg/display/theme"
	"github.com/ygelfand/libcountertop/pkg/display/ui"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/startup"
	"github.com/ygelfand/libcountertop/pkg/say"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/gpu"
	hwtouch "github.com/ygelfand/echolocal/internal/hardware/touch"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/ui/reveal"
)

func init() {
	component.Register(
		sharedcomponent.Hardware,
		Get,
		sharedcomponent.Order(5),
		component.Needs(board.Panel),
	)
}

type Splash struct {
	hold     *panel.Claim
	finished atomic.Bool
}

func (s *Splash) Finished() bool { return s.finished.Load() }

var (
	once   sync.Once
	shared *Splash
)

func Get() *Splash { once.Do(func() { shared = &Splash{} }); return shared }

func (s *Splash) Name() string { return "splash" }

func (s *Splash) Start(context.Context) error {
	s.finished.Store(false)
	s.hold = display.Get().Claim(panel.PriorityBoot)
	s.hold.Show(func(p *panel.Panel) error {
		surface := ui.Of(p)
		w, h := surface.Size()
		ui.Clear(surface, ui.Rect{W: w, H: h})
		return nil
	})
	return nil
}

func (s *Splash) Run(ctx context.Context) error {
	if s.hold == nil {
		<-ctx.Done()
		return nil
	}

	defer s.hold.Release()
	w, h := bootview.Viewed(ctx, display.Get())

	completed := bootview.Splash{
		Claim:    s.hold,
		Display:  display.Get(),
		Layers:   gpu.Get(),
		Contacts: &hwtouch.Get().Contacts,
		Reveal:   reveal.New(chosen()),
		Theme:    chosen,
		Version:  layout.Version,
		Progress: func() []startup.Progress { return component.Default().Progress() },
		Ready:    handOver,
		Raster: func(p *panel.Panel, progress []startup.Progress) error {
			bootview.Raster(ui.Of(p), chosen(), progress, layout.Version, say.T("boot.skip"))
			return nil
		},
		Blank: func(p *panel.Panel) error {
			ui.Fill(ui.Of(p), chosen().Background)
			return nil
		},
	}.Play(ctx, w, h)
	if completed {
		s.finished.Store(true)
	}
	return nil
}

func handOver() bool { return component.Default().Ready() }

func chosen() theme.Theme {
	s := config.Get().Screen
	t, ok := theme.ByName(style.Theme(s.Style, s.Theme))
	if !ok {
		return theme.Default()
	}
	return t
}
