package assistant

import (
	"sync"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/feature/voice"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/touch"
	sharedscreen "github.com/ygelfand/libcountertop/pkg/display/assistant"
)

func init() {
	component.Register(component.Device, Get, component.Order(45), component.Needs(board.Panel))
}

var get = sync.OnceValue(func() *sharedscreen.Assistant {
	a := sharedscreen.New(sharedscreen.Dependencies{Microphone: mic.Get(), Controller: voice.Get()})
	voice.Shown.Listen(a.Show)
	touch.Contacts.Listen(func(c touch.Contact) {
		if c.Phase == touch.Down {
			a.Cancel(c.ID)
		}
	})
	return a
})

func Get() *sharedscreen.Assistant { return get() }
