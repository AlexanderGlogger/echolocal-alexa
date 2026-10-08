package homecontrol

import (
	"sync"

	shareddashboard "github.com/ygelfand/libcountertop/pkg/display/dashboard"
	sharedhome "github.com/ygelfand/libcountertop/pkg/homeassistant/homecontrol"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/states"
)

var get = sync.OnceValue(func() *sharedhome.Engine {
	e := sharedhome.New(
		sharedhome.Dependencies{
			Settings: config.HomeSection,
			Client:   homeassistant.Get(),
			States:   states.Get(),
			Shell:    shell.Get(),
		},
	)
	component.Settings.Add(e.Controls())
	return e
})

func Get() *sharedhome.Engine { return get() }
func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(6),
		component.Needs(board.Panel),
	)
	dashboard.Tabs().Add(10, func() []shareddashboard.Tab { return Get().DashboardTabs() })
}
