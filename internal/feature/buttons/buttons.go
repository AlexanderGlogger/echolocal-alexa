package buttons

import (
	"log/slog"
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
)

func init() {
	component.Register(sharedcomponent.Device, Get)
}

const (
	Press = "press"
	Hold  = "hold"
)

var order = []buttons.Name{buttons.Action, buttons.Mute, buttons.VolumeDown, buttons.VolumeUp}

type Events struct {
	events map[buttons.Name]*esphome.Event
}

var (
	once   sync.Once
	shared *Events
)

func Get() *Events {
	once.Do(func() {
		shared = &Events{events: map[buttons.Name]*esphome.Event{}}
		for _, name := range order {
			shared.events[name] = &esphome.Event{
				Base: esphome.Base{
					ObjectID: "button_" + string(name),
					Name:     label(string(name)) + " button",
				},
				Types: []string{Press, Hold},
			}
		}
		buttons.Get().Events.Listen(shared.report)
	})
	return shared
}

func (e *Events) Name() string { return "button events" }

func (e *Events) Entities() []esphome.Entity {
	out := make([]esphome.Entity, 0, len(order))
	for _, name := range order {
		out = append(out, e.events[name])
	}
	return out
}

func (e *Events) report(ev buttons.Event) {
	if ev.Kind == buttons.Repeat {
		return
	}
	slog.Debug("button", "button", ev.Name, "kind", ev.Kind)

	entity, ok := e.events[ev.Name]
	if !ok {
		return
	}
	if ev.Kind == buttons.Hold {
		entity.Trigger(Hold)
		return
	}
	entity.Trigger(Press)
}

func label(name string) string {
	out := []rune(name)
	for i, r := range out {
		if r == '_' {
			out[i] = ' '
		}
	}
	out[0] -= 32
	return string(out)
}
