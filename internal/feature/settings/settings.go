//go:build board_checkers || board_cronos || board_rook

package settings

import (
	"slices"
	"sync"

	"github.com/ygelfand/libcountertop/pkg/say"

	gogui "github.com/go-gui-org/go-gui/gui"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/drawer"
	"github.com/ygelfand/echolocal/internal/feature/shell"
)

func init() {
	component.Register(component.Device, Get, component.Order(36), component.Needs(board.Panel))
}

type Settings struct{}

var (
	once   sync.Once
	shared *Settings
)

func Get() *Settings {
	once.Do(func() {
		shared = &Settings{}

		say.Use(config.Get().Screen.Language)

		drawer.Get().Add(drawer.Entry{
			Name:  func() string { return say.T("settings.title") },
			Order: drawer.OrderSettings,
			Glyph: func() string { return gogui.IconGear },
			Open:  Open,
		})
	})
	return shared
}

func (s *Settings) Name() string { return "settings" }

func Open() { shell.Get().Push(root()) }

var pages = map[string]func() *shell.Page{"settings": root}

func OpenPage(name string) bool {
	page, ok := pages[name]
	if ok {
		shell.Get().Push(page())
	}
	return ok
}

func Pages() []string {
	out := make([]string, 0, len(pages))
	for name := range pages {
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}
