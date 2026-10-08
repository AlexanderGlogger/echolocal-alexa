package splash

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	bootview "github.com/ygelfand/libcountertop/pkg/display/boot"
	"github.com/ygelfand/libcountertop/pkg/display/style"
	"github.com/ygelfand/libcountertop/pkg/runtime/startup"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/gpu"
	"github.com/ygelfand/echolocal/internal/hardware/touch"
	"github.com/ygelfand/echolocal/internal/ui"
	"github.com/ygelfand/echolocal/internal/ui/reveal"
	"github.com/ygelfand/echolocal/internal/ui/theme"
)

func init() {
	component.Register(component.Hardware, Get, component.Order(5), component.Needs(board.Panel))
}

const frame = startup.Frame

type Splash struct {
	hold     *display.Claim
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
	s.hold = display.Get().Claim(display.PriorityBoot)
	s.hold.Show(func(p *display.Panel) error {
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
	palette := chosen()
	w, h := viewed(ctx)
	rv := reveal.New(palette)
	layer, err := gpu.Get().Open(w, h)
	if err != nil {
		slog.Error(
			"the boot animation could not start",
			"err",
			err,
			"size",
			fmt.Sprintf("%dx%d", w, h),
		)
	} else {
		defer layer.Close()
		if err := layer.Place(ui.Rect{W: w, H: h}); err != nil {
			slog.Error("placing the boot animation", "err", err)
		}
	}

	var skipped, offering atomic.Bool
	skip := bootview.Skip(w, h)
	stopSkip := touch.Contacts.Listen(func(c touch.Contact) {
		if c.Phase == touch.Down && offering.Load() && skip.Contains(c.X, c.Y) {
			slog.Info("boot screen skipped")
			skipped.Store(true)
		}
	})
	defer stopSkip()

	scene := openBootScene(w, h)
	if scene != nil {
		defer func() {
			if scene != nil {
				scene.Close()
			}
		}()
	}
	fresh := true
	completed := startup.Run(ctx, startup.Hooks{
		Progress: func() []startup.Progress { return component.Default().Progress() },
		Ready:    handOver, Skipped: skipped.Load, OfferSkip: offering.Store,
		Draw: func(progress []startup.Progress) {
			if scene != nil {
				if err := scene.Draw(progress, true); err == nil {
					return
				} else {
					slog.Error("drawing boot progress", "err", err)
					scene.Close()
					scene = nil
				}
			}
			s.hold.Show(
				func(p *display.Panel) error { drawBoot(p, chosen(), progress); return nil },
			)
		},
		Changed: func(text string) { slog.Info("coming up", "waiting", text) },
		Animate: func(m startup.Moment) bool {
			if layer == nil {
				return false
			}
			if err := rv.Shade(
				layer,
				fresh,
				w,
				h,
				reveal.Moment(m),
			); err != nil {
				slog.Error("the boot animation stopped", "err", err)
				_ = layer.Close()
				layer = nil
				return false
			}
			fresh = false
			return true
		},
	})
	if !completed {
		return nil
	}
	s.finished.Store(true)

	s.hold.Show(func(p *display.Panel) error {
		ui.Fill(ui.Of(p), chosen().Background)
		return nil
	})
	time.Sleep(2 * frame)

	slog.Info("ready")
	if scene != nil {
		scene.Close()
		scene = nil
	}
	return nil
}

func viewed(ctx context.Context) (int, int) {
	for range 40 {
		fw, fh := display.Get().Native()
		if w, h := display.Get().Orientation().Size(fw, fh); w > 0 && h > 0 {
			return w, h
		}
		if !wait(ctx, 50*time.Millisecond) {
			break
		}
	}
	return 0, 0
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

func wait(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
