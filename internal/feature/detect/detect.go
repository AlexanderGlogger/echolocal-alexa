package detect

import (
	"context"
	"log/slog"
	"sync"
	"time"

	sharedengine "github.com/ygelfand/libcountertop/pkg/inference/detect"

	"github.com/ygelfand/echolocal/internal/layout"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/diag"
	"github.com/ygelfand/echolocal/internal/feature/voice"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/hardware/led"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/lib/wake"
	"github.com/ygelfand/echolocal/internal/service"
)

func init() {
	component.Register(component.Device, Get, component.Order(40),
		component.Supervise(service.Restart(time.Second, 30*time.Second)))
}

type Detect struct {
	engine *Engine
	busy   *wakeBusy
	stop   *sharedengine.StopWord
}

var (
	once   sync.Once
	shared *Detect
)

func Get() *Detect {
	once.Do(func() { shared = newDetect() })
	return shared
}

var pauses []func() bool

func PauseWhen(f func() bool) { pauses = append(pauses, f) }

func paused() bool {
	for _, f := range pauses {
		if f() {
			return true
		}
	}
	return false
}

func newDetect() *Detect {
	e := New(StopSlot+1, mic.Get())
	e.Paused = paused

	e.Threshold = func(slot int) float64 {
		if slot == StopSlot {
			return config.Get().Wake.Stop.Threshold
		}
		return wakeword.Threshold(slot)
	}

	e.OnDetect = func(slot int) {
		if slot == StopSlot {
			voice.Get().Interrupt()
			return
		}
		voice.Get().Start(slot)
	}

	d := &Detect{engine: e, busy: newWakeBusy(led.Get().Busy(), e.Ready)}
	d.stop = sharedengine.NewStopWord(e, sharedengine.StopOptions{
		Directory: layout.StateDir, DeviceID: component.DeviceMicrophone,
		Read: func() float64 { return config.Get().Wake.Stop.Threshold },
		Save: func(v float64) error { return config.Set().Stop().Threshold(v) },
	})
	e.OnReady = d.busy.scored

	// Home Assistant pushes a wake word selection only when the user changes one.
	e.Load = func() error {
		turn := voice.Get()
		turn.SetSlots(d.load(turn.Slots()))
		d.stop.Load()
		return nil
	}

	voice.Get().OnWakeWord(d.load, diag.Get().Measure)

	ours := wake.Lib().Ours()
	slog.Info("wake words installed", "count", len(ours),
		"openwakeword", len(wake.OfKind(ours, wake.KindOpenWakeWord)),
		"microwakeword", len(wake.OfKind(ours, wake.KindMicroWakeWord)))

	return d
}

func (d *Detect) Name() string { return "wake" }

func (d *Detect) Start(ctx context.Context) error { return d.engine.Start(ctx) }

func (d *Detect) Run(ctx context.Context) error { return d.engine.Run(ctx) }

func (d *Detect) Close() error { return d.engine.Close() }

func (d *Detect) load(ids []string) []string {
	d.busy.begin()

	models := wake.Lib().Ensure(ids)

	accepted := make([]string, wakeword.Slots)
	var loaded []int
	for slot := range wakeword.Slots {
		if slot >= len(ids) || ids[slot] == "" {
			d.engine.Clear(slot)
			continue
		}

		m, ok := wake.Find(models, ids[slot])
		if !ok {
			slog.Warn("unknown wake word", "slot", slot+1, "id", ids[slot])
			d.engine.Clear(slot)
			continue
		}
		if err := d.engine.Use(slot, m); err != nil {
			slog.Error(
				"loading the selected wake word failed",
				"slot",
				slot+1,
				"id",
				m.ID,
				"err",
				err,
			)
			d.engine.Clear(slot)
			continue
		}
		accepted[slot] = m.ID
		loaded = append(loaded, slot)
	}

	d.busy.waitFor(loaded)
	return accepted
}
