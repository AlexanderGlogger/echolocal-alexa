package privacy

import (
	"context"
	"log/slog"
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/libcountertop/pkg/hook"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	setting "github.com/ygelfand/libcountertop/pkg/settings"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/mute"
	"github.com/ygelfand/echolocal/internal/feature/screen"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(35),
		component.Needs(board.Panel),
	)
}

type Privacy struct {
	Changed hook.Hook[Marks]

	camera *esphome.BinarySensor

	mu    sync.Mutex
	marks Marks
}

var (
	once   sync.Once
	shared *Privacy
)

func Get() *Privacy {
	once.Do(func() {
		shared = &Privacy{
			camera: &esphome.BinarySensor{
				Base: esphome.Base{
					ObjectID: "camera_covered",
					Name:     "Camera covered",
					Icon:     "mdi:camera-off",
				},
			},
		}
		buttons.Get().Shutter.Listen(shared.covered)
		mute.Get().Changed.Listen(shared.muted)
		screen.Get().Restyled.Listen(func(schema.Screen) { shared.show() })
	})
	return shared
}

func (p *Privacy) Name() string { return "privacy" }

func (p *Privacy) Entities() []esphome.Entity {
	if !component.Board().Has(board.Shutter) {
		return nil
	}
	return []esphome.Entity{p.camera}
}

func (p *Privacy) Restore(config.Config) { p.show() }

func (p *Privacy) SetMarks(on bool) {
	if err := screen.Get().Set("marks", setting.OnOff(on)); err != nil {
		slog.Error("saving the privacy marks setting failed", "err", err)
	}
}

func (p *Privacy) Start(context.Context) error {
	if component.Board().Has(board.Shutter) {
		p.covered(buttons.Get().Covered())
	}

	muted, err := mute.Get().Muted()
	if err != nil {
		slog.Warn("cannot tell whether the microphones are cut", "err", err)
	}
	p.muted(muted)
	return nil
}

func (p *Privacy) MicMuted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.marks.MicMuted
}

func (p *Privacy) CameraCovered() bool { return p.camera.Get() }

func (p *Privacy) covered(on bool) {
	p.camera.Set(on)

	p.mu.Lock()
	p.marks.CameraBlocked = on
	p.mu.Unlock()

	p.show()
}

func (p *Privacy) muted(on bool) {
	p.mu.Lock()
	p.marks.MicMuted = on
	p.mu.Unlock()

	p.show()
}

func (p *Privacy) Current() Marks {
	if !config.Get().Screen.Marks {
		return Marks{}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.marks
}

func (p *Privacy) show() { p.Changed.Emit(p.Current()) }

func (p *Privacy) Watch(changed func()) func() {
	return p.Changed.Listen(func(Marks) { changed() })
}
