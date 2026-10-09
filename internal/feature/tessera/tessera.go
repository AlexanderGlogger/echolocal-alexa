package tessera

import (
	"path/filepath"
	"sync"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"
	shared "github.com/ygelfand/libcountertop/pkg/tessera"
	tesseraview "github.com/ygelfand/libcountertop/pkg/tessera/widgets"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/homecontrol"
	"github.com/ygelfand/echolocal/internal/feature/idle"
	"github.com/ygelfand/echolocal/internal/feature/message"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/layout"
)

func init() {
	component.Register(
		sharedcomponent.Network,
		Get,
		sharedcomponent.Order(95),
		component.Needs(board.Panel),
	)
}

var get = sync.OnceValue(func() *shared.Set {
	home := homecontrol.Get()
	alert, unalert := tesseraview.Alerter(message.Get())
	front := func() {
		shell.Get().Close()
		idle.Get().Wake()
		dashboard.Tabs().Clock()
	}
	set := shared.NewSet(shared.New(shared.Options{
		Facing:   shared.Dashboard,
		Device:   component.DeviceDashboard,
		Board:    func() string { return component.Board().Tessera.Board },
		Size:     display.Get().Size,
		Grid:     home.Grid,
		Node:     func() string { return layout.Slug(config.Get().Device.Name) },
		Name:     func() string { return config.Get().Device.Name },
		Speaking: home.Tessera,
		Keep:     filepath.Join(layout.TesseraDir, "tessera-dashboard.json"),
		Alert:    alert,
		Unalert:  unalert,
		Front:    front,
	}))
	home.Speaking.Listen(func(schema.HomeMode) { component.Reconnect.Emit(struct{}{}) })
	home.Regrid.Listen(func(schema.TesseraGrid) { set.Regrid() })
	display.Get().Attached.Listen(func(struct{}) { set.Regrid() })
	set.Changed.Listen(func(struct{}) { shell.Get().Redraw() })
	set.Told.Listen(component.Fire.Emit)
	return set
})

func Get() *shared.Set { return get() }
