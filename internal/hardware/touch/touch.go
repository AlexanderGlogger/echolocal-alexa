package touch

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ygelfand/libcountertop/pkg/hook"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/lib/input"
	"github.com/ygelfand/echolocal/internal/service"
)

func init() {
	component.Register(component.Hardware, Get, component.Order(30),
		component.Needs(board.Panel),
		component.Supervise(service.Restart(time.Second, time.Minute)))
}

const Name = "goodix-ts"

var Contacts hook.Hook[Contact]

var Gestures hook.Hook[Gesture]

type Touch struct {
	mu   sync.Mutex
	dev  *input.Device
	rec  *Recognizer
	last time.Time
}

var started = time.Now()

func (t *Touch) Since() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.last.IsZero() {
		return time.Since(started)
	}
	return time.Since(t.last)
}

func (t *Touch) recognize(c Contact) (Gesture, bool) {
	t.mu.Lock()
	if t.rec == nil {
		t.rec = NewRecognizer(display.Get().Size())
	}
	rec := t.rec
	t.mu.Unlock()

	return rec.Feed(c)
}

var (
	once   sync.Once
	shared *Touch
)

func Get() *Touch { once.Do(func() { shared = &Touch{} }); return shared }

func (t *Touch) Name() string { return "touch" }

func (t *Touch) Start(context.Context) error {
	d, err := find()
	if err != nil {
		slog.Error("the touchscreen would not open", "err", err)
		return nil
	}

	t.mu.Lock()
	t.dev = d
	t.mu.Unlock()

	slog.Info("touchscreen", "name", d.Name, "path", d.Path)
	return nil
}

func (t *Touch) Startup() component.Progress {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.dev == nil {
		return component.Progress{Failed: true, Doing: "no touchscreen"}
	}
	return component.Progress{Done: true}
}

func (t *Touch) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.dev == nil {
		return nil
	}
	err := t.dev.Close()
	t.dev = nil
	return err
}

func (t *Touch) Run(ctx context.Context) error {
	t.mu.Lock()
	d := t.dev
	t.mu.Unlock()

	if d == nil {
		<-ctx.Done()
		return nil
	}

	failed := make(chan error, 1)
	go func() { failed <- t.read(d) }()

	// A read on the event node blocks until the next touch; only Close ends it.
	select {
	case <-ctx.Done():
		return nil
	case err := <-failed:
		return err
	}
}

func (t *Touch) read(d *input.Device) error {
	dec := decoder{place: place}
	for {
		e, err := d.Read()
		if err != nil {
			return fmt.Errorf("touch: reading %s: %w", d.Path, err)
		}
		for _, c := range dec.event(e, time.Now()) {
			t.Feed(c)
		}
	}
}

func (t *Touch) Feed(c Contact) {
	t.mu.Lock()
	t.last = time.Now()
	t.mu.Unlock()

	Contacts.Emit(c)

	if g, ok := t.recognize(c); ok {
		Gestures.Emit(g)
	}
}

func place(px, py int) (int, int) {
	fw, fh := display.Get().Native()
	if fw == 0 {
		unplaced.Do(func() {
			slog.Warn("no panel, so touches are reported where the hardware put them")
		})
		return px, py
	}
	return display.Get().Orientation().Unproject(fw, fh, px, py)
}

var unplaced sync.Once

func find() (*input.Device, error) {
	devices, err := input.List()
	if err != nil {
		return nil, err
	}

	var found *input.Device
	for _, d := range devices {
		if d.Name == Name {
			found = d
			continue
		}
		_ = d.Close()
	}
	if found == nil {
		return nil, fmt.Errorf("touch: no input device called %q", Name)
	}
	return found, nil
}
