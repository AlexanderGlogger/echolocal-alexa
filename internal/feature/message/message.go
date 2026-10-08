package message

import (
	"sync"

	shared "github.com/ygelfand/libcountertop/pkg/display/message"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/touch"
)

var get = sync.OnceValue(func() *shared.Messages {
	m := shared.New()
	touch.Contacts.Listen(func(c touch.Contact) {
		if c.Phase == touch.Down {
			m.Dismiss(c.ID)
		}
	})
	return m
})

func Get() *shared.Messages { return get() }
func init() {
	component.Register(component.Device, Get, component.Order(40), component.Needs(board.Panel))
}
