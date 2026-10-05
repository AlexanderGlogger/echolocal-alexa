// Package splash is what the panel shows while the device is coming up: the mark, then what it is
// still waiting for.
package splash

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/gpu"
	"github.com/ygelfand/echolocal/internal/hardware/touch"
	"github.com/ygelfand/echolocal/internal/ui"
	"github.com/ygelfand/echolocal/internal/ui/reveal"
	"github.com/ygelfand/echolocal/internal/ui/style"
	"github.com/ygelfand/echolocal/internal/ui/theme"
)

func init() {
	// First of the hardware, for the same reason the ring is: it is the only thing the device can
	// say before anything else works.
	component.Register(component.Hardware, Get, component.Order(5), component.Needs(board.Panel))
}

const (
	// logoFor is how long the mark is up before the screen starts saying what it is waiting for.
	// Long enough to be a greeting, short enough that a stuck device says so quickly.
	logoFor = 3 * time.Second

	// listFor is the least the list stays up once it has appeared, so a device that comes up in a
	// second does not flash it past unread.
	listFor = 3 * time.Second

	settle = 500 * time.Millisecond
)

const (
	frame    = time.Second / 60
	traceBy  = 2000 * time.Millisecond
	traceFor = 1400 * time.Millisecond
	stepFor  = 350 * time.Millisecond
	moveFor  = 700 * time.Millisecond
	finish   = 1100 * time.Millisecond
)

type Splash struct{ hold *display.Claim }

var (
	once   sync.Once
	shared *Splash
)

func Get() *Splash { once.Do(func() { shared = &Splash{} }); return shared }

func (s *Splash) Name() string { return "splash" }

// Start takes the panel before anything else has run, clear so the reveal underneath shows through.
func (s *Splash) Start(context.Context) error {
	s.hold = display.Get().Claim(display.PriorityBoot)
	s.hold.Show(func(p *display.Panel) error {
		surface := ui.Of(p)
		w, h := surface.Size()
		ui.Clear(surface, ui.Rect{W: w, H: h})
		return nil
	})
	return nil
}

// Run plays the reveal, then what the device is still waiting for, and finishes once everything
// is up.
func (s *Splash) Run(ctx context.Context) error {
	if s.hold == nil {
		<-ctx.Done()
		return nil
	}

	palette := chosen()
	w, h := viewed(ctx)
	rv := reveal.New(palette)
	layer, err := gpu.Open(w, h)
	if err != nil {
		slog.Error("the boot animation could not start", "err", err, "size", fmt.Sprintf("%dx%d", w, h))
	} else {
		defer layer.Close()
		if err := layer.Place(ui.Rect{W: w, H: h}); err != nil {
			slog.Error("placing the boot animation", "err", err)
		}
	}

	var skipped, offering atomic.Bool
	skip := skipBox(w, h)
	stopSkip := touch.Contacts.Listen(func(c touch.Contact) {
		if c.Phase == touch.Down && offering.Load() && skip.Contains(c.X, c.Y) {
			slog.Info("boot screen skipped")
			skipped.Store(true)
		}
	})
	defer stopSkip()

	began := time.Now()
	tick := time.NewTicker(frame)
	defer tick.Stop()

	var m reveal.Moment
	var listed, done, finishing, stepped time.Time
	shown := 0
	waited := false
	said := "\x00"
	fresh := true

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}
		now := time.Now()
		m.At = now.Sub(began)
		progress := component.Default().Progress()

		if m.At > traceBy {
			if shown < up(progress) && now.Sub(stepped) >= stepFor {
				shown++
				stepped = now
			}
			want := 1.0
			if len(progress) > 0 && !skipped.Load() {
				want = float64(shown) / float64(len(progress))
			}
			m.Trace = math.Min(want, m.Trace+frame.Seconds()/traceFor.Seconds())
		}

		if m.At >= logoFor {
			m.Header = math.Min(1, m.Header+frame.Seconds()/moveFor.Seconds())
		}
		if m.Header >= 1 {
			if listed.IsZero() {
				listed = now
			}
			offering.Store(!settled(progress))
			if text := summary(progress); text != said {
				said = text
				s.hold.Show(func(p *display.Panel) error {
					drawBoot(p, chosen(), progress)
					return nil
				})
				slog.Info("coming up", "waiting", text)
			}
		}

		if !listed.IsZero() && !settled(progress) {
			waited = true
		}
		if !handOver() {
			done = time.Time{}
		} else if done.IsZero() {
			done = now
		}
		if finishing.IsZero() && !listed.IsZero() && !done.IsZero() && now.Sub(done) >= settle && (!waited || now.Sub(listed) >= listFor) && (settled(progress) || skipped.Load()) {
			finishing = now
		}
		if !finishing.IsZero() && m.Trace >= 1 {
			m.Ready = math.Min(1, m.Ready+frame.Seconds()/finish.Seconds())
		}

		if layer != nil {
			if err := rv.Shade(layer, fresh, w, h, m); err != nil {
				slog.Error("the boot animation stopped", "err", err)
				layer.Close()
				layer = nil
			}
			fresh = false
		}
		if m.Ready >= 1 || (layer == nil && !finishing.IsZero()) {
			break
		}
	}
	s.hold.Show(func(p *display.Panel) error {
		ui.Fill(ui.Of(p), chosen().Background)
		return nil
	})
	time.Sleep(2 * frame)

	slog.Info("ready")
	s.hold.Release()
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

func up(progress []component.Progress) int {
	n := 0
	for _, p := range progress {
		if p.Settled() {
			n++
		}
	}
	return n
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

// summary is what the screen currently says, for deciding whether to draw it again.
func summary(progress []component.Progress) string {
	var s string
	for _, p := range progress {
		s += fmt.Sprintf("%s:%v:%s ", p.Name, p.Settled(), p.Doing)
	}
	return s
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
