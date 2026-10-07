//go:build board_checkers || board_cronos || board_rook

package media

import (
	"sync"

	gogui "github.com/go-gui-org/go-gui/gui"
	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/drawer"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	sharedcard "github.com/ygelfand/libcountertop/pkg/display/mediacard"
	"github.com/ygelfand/libcountertop/pkg/say"
)

type screen = sharedcard.Screen

var cardOnce sync.Once
var theCard *sharedcard.Card

func card() *sharedcard.Card {
	cardOnce.Do(func() {
		theCard = sharedcard.New(
			sharedcard.Dependencies{Player: Get(), Idle: config.IdleSection, Shell: shell.Get()},
		)
	})
	return theCard
}
func Page() shell.View { return card().Page() }
func Showing() bool    { return shell.Get().Top() == Page() }
func onRail() {
	drawer.Get().
		Add(drawer.Entry{Name: func() string { return say.T("rail.media") }, Order: drawer.OrderPlayer, Glyph: func() string { return gogui.IconMusic }, Open: func() {
			if now := Get().Now(); now.Playing || now.Paused {
				card().Open()
				return
			}
			shell.Get().Push(Page())
		}})
}

func init() {
	component.Register(
		component.Device,
		func() *sharedcard.Card { c := card(); onRail(); Get().Begun.Listen(c.Began); return c },
		component.Order(39),
		component.Needs(board.Panel),
	)
}
func (p *Player) Open() { card().Open() }
