package touch

import (
	"sync"
	"time"

	"github.com/ygelfand/libcountertop/pkg/input/touch"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/service"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/display"
)

func init() {
	component.Register(sharedcomponent.Hardware, Get, sharedcomponent.Order(30),
		component.Needs(board.Panel),
		sharedcomponent.Supervise(service.Restart(time.Second, time.Minute)))
}

const device = "goodix-ts"

var get = sync.OnceValue(func() *touch.Screen { return touch.New(device, display.Get()) })

func Get() *touch.Screen { return get() }
