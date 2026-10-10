package control

import (
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/hardware/buttons"
)

func (c *Control) inputEngine() *harness.Input {
	c.inputOnce.Do(func() { c.input = newInput() })
	return c.input
}
func newInput() *harness.Input {
	options := harness.InputOptions{}
	options.Buttons = []string{"up", "down", "mute", "action"}
	names := map[string]buttons.Name{
		"up":     buttons.VolumeUp,
		"down":   buttons.VolumeDown,
		"mute":   buttons.Mute,
		"action": buttons.Action,
	}
	options.Press = func(name string) error {
		buttons.Get().Press(buttons.Event{Name: names[name], Kind: buttons.Tap})
		return nil
	}
	configureTouch(&options)
	return harness.NewInput(options)
}
