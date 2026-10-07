package assistant

import (
	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/feature/voice"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/touch"
	sharedscreen "github.com/ygelfand/libcountertop/pkg/display/assistant"
	"sync"
)

type Assistant = sharedscreen.Assistant

func init() {
	component.Register(component.Device, Get, component.Order(45), component.Needs(board.Panel))
}

var get = sync.OnceValue(func() *Assistant {
	a := sharedscreen.New(
		sharedscreen.Options{
			Level: func() float64 { return mic.Get().Level() },
			Stop:  func() bool { return voice.Get().Stop() },
		},
	)
	voice.Shown.Listen(a.Show)
	touch.Contacts.Listen(func(c touch.Contact) {
		if c.Phase == touch.Down {
			a.Cancel(c.ID)
		}
	})
	return a
})

func Get() *Assistant { return get() }
