package display

import (
	"sync"
	"time"

	"github.com/ygelfand/libcountertop/pkg/display/geometry"
	"github.com/ygelfand/libcountertop/pkg/display/panel"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/service"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/layout"
)

const socket = "/dev/socket/echolocal-surface"

func init() {
	component.Register(
		sharedcomponent.Hardware,
		Get,
		sharedcomponent.Order(4),
		component.Needs(board.Panel),
		sharedcomponent.Supervise(service.Restart(time.Second, 30*time.Second)),
	)
}

func Mounted() geometry.Orientation { return geometry.Orientation(component.Board().PanelRotation) }

func config() panel.Config {
	return panel.Config{
		Framebuffer: layout.FBDevice,
		Socket:      socket,
		Mounted:     Mounted(),
	}
}

var get = sync.OnceValue(func() *panel.Driver { return panel.NewDriver(config) })

func Get() *panel.Driver { return get() }
