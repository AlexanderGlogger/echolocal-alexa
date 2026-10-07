package discovery

import (
	"strings"
	"sync"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dashboard"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/web"
	"github.com/ygelfand/echolocal/internal/hardware/wifi"
	"github.com/ygelfand/echolocal/internal/layout"
	shared "github.com/ygelfand/libcountertop/pkg/network/discovery"
	sharedpeer "github.com/ygelfand/libcountertop/pkg/network/peer"
	"github.com/ygelfand/libcountertop/pkg/say"
)

const TabKind = "call"

type identity struct{}

func (identity) Self() sharedpeer.Peer {
	b := component.Board()
	return sharedpeer.Peer{
		ID:      ID(),
		Name:    config.Get().Device.Name,
		Model:   b.Model,
		Board:   b.Codename,
		Caps:    caps(),
		Project: strings.ToLower(layout.Manufacturer), Version: layout.Version, Port: web.Port, Host: layout.Slug(config.Get().Device.Name),
	}
}

func ID() string { return strings.ToLower(strings.ReplaceAll(wifi.Get().MAC(), ":", "")) }
func caps() []string {
	c := []string{"audio"}
	if component.Board().CameraWidth > 0 {
		c = append(c, "video")
	}
	return c
}

var get = sync.OnceValue(func() *shared.Discovery {
	d := shared.New(identity{})
	d.Changed.Listen(func(struct{}) { shell.Get().Redraw() })
	return d
})

func Get() *shared.Discovery { return get() }
func init() {
	component.Register(component.Network, Get, component.Order(60), component.Needs(board.Panel))
	dashboard.AddTabs(100, func() []dashboard.Tab { return []dashboard.Tab{{Kind: TabKind, Key: TabKind, Name: say.T("call.tab")}} })
}
