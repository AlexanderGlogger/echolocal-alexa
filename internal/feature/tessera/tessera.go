package tessera

import (
	"path/filepath"
	"sync"

	"github.com/ygelfand/libcountertop/pkg/display/theme"
	"github.com/ygelfand/libcountertop/pkg/display/widgets"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"
	"github.com/ygelfand/libcountertop/pkg/settings/timezone"
	shared "github.com/ygelfand/libcountertop/pkg/tessera"
	"github.com/ygelfand/libcountertop/pkg/tessera/firmwarerelay"
	"github.com/ygelfand/libcountertop/pkg/tessera/firmwareview"
	tesseraview "github.com/ygelfand/libcountertop/pkg/tessera/widgets"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	"github.com/ygelfand/echolocal/internal/feature/homecontrol"
	"github.com/ygelfand/echolocal/internal/feature/idle"
	"github.com/ygelfand/echolocal/internal/feature/message"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/states"
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
	component.Register(
		sharedcomponent.Network,
		Firmware,
		sharedcomponent.Order(96),
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
		Speaking: func() bool { return home.Tessera() && !home.Firmware() },
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
	firmware()
	return set
})

func Get() *shared.Set { return get() }

var firmware = sync.OnceValue(func() *firmwareview.View {
	v := firmwareview.New(firmwareview.Options{
		Helper:   layout.TesseraHelper,
		Screen:   screen(),
		Panel:    display.Get().Size,
		Diagonal: component.Board().Diagonal,
		Language: func() string { return config.Get().Screen.Language },
		Grid: func() (int, int) {
			g := homecontrol.Get().Grid()
			return g.Columns, g.Rows
		},
		Theme:   func() theme.Theme { return widgets.Palette(config.Get().Screen) },
		Themed:  homecontrol.Get().Themed,
		Turn:    homecontrol.Get().Turn,
		Bottom:  homecontrol.Get().Bottom,
		Follow:  homecontrol.Get().Follow,
		Product: layout.Manufacturer,
		Board:   component.Board().Tessera.Board,
		Redraw:  shell.Get().Redraw,
	})
	home := homecontrol.Get()
	home.Redrawn.Listen(func(on bool) {
		v.Set(on)
		component.Reconnect.Emit(struct{}{})
	})
	display.Get().Attached.Listen(func(struct{}) { v.Set(home.Firmware()) })
	v.Set(home.Firmware())
	return v
})

func Firmware() *firmwareview.View { return firmware() }

func screen() func(upright bool) firmwareview.Screen {
	dashboard := firmwareview.Screen{
		Dir: layout.TesseraFirmwareDir,
		Relay: firmwarerelay.New(firmwarerelay.Options{
			Word:    "dashboard",
			Device:  component.DeviceDashboard,
			Port:    layout.TesseraPort,
			Node:    func() string { return layout.Slug(config.Get().Device.Name) },
			Zone:    func() string { return timezone.Spec(config.Get().Time) },
			Home:    homeassistant.Get(),
			States:  states.Get(),
			Changed: func() { component.Reconnect.Emit(struct{}{}) },
		}),
	}
	return func(bool) firmwareview.Screen { return dashboard }
}
