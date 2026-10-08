//go:build board_checkers || board_cronos

package call

import (
	"sync"

	sharedcall "github.com/ygelfand/libcountertop/pkg/media/call"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/detect"
	"github.com/ygelfand/echolocal/internal/feature/discovery"
	"github.com/ygelfand/echolocal/internal/feature/livecam"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/message"
	"github.com/ygelfand/echolocal/internal/feature/privacy"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/web"
	"github.com/ygelfand/echolocal/internal/hardware/display"
)

type componentCall struct{ *sharedcall.Calls }

func (c *componentCall) Restore(v config.Config) { c.Calls.Restore(v.Call) }

var once sync.Once
var shared *sharedcall.Calls

func Get() *sharedcall.Calls {
	once.Do(func() {
		shared = sharedcall.New(sharedcall.Dependencies{
			Settings: config.CallSection,
			Peers:    discovery.Get(),
			Port:     web.Port,
			Shell:    shell.Get(),
			Privacy:  privacy.Get(),
			Media:    media.Get(),
			Sound:    device{},
			Camera:   sharedcall.NewCamera(livecam.Get().Sessions()),
			Display:  display.Get(),
			Messages: message.Get(),
		})
		component.Settings.Add(shared.Controls())
	})
	return shared
}

func init() {
	component.Register(
		sharedcomponent.Network,
		func() *componentCall { return &componentCall{Get()} },
		sharedcomponent.Order(62),
		component.Needs(board.Panel),
	)
	web.Handle("POST /call/offer", Get().OfferHandler())
	web.Handle("POST /call/answer", Get().AnswerHandler())
	web.Handle("POST /call/end", Get().EndHandler())
	detect.PauseWhen(Get().PausesWake)
}
