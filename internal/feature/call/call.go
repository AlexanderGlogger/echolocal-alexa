package call

import (
	"sync"

	sharedcall "github.com/ygelfand/libcountertop/pkg/media/call"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/detect"
	"github.com/ygelfand/echolocal/internal/feature/discovery"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/tones"
	"github.com/ygelfand/echolocal/internal/feature/web"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
)

type componentCall struct{ *sharedcall.Calls }

func (c *componentCall) Restore(v config.Config) { c.Calls.Restore(v.Call) }

var once sync.Once
var shared *sharedcall.Calls

func Get() *sharedcall.Calls {
	once.Do(func() {
		d := sharedcall.Dependencies{
			Settings: config.CallSection,
			Peers:    discovery.Get(),
			Port:     web.Port,
			Media:    media.Get(),
			Sound:    device{},
			Tones:    tones.Get(),
		}
		front(&d)
		shared = sharedcall.New(d)
		configureSettings(shared)
		component.Settings.Add(shared.Controls())
	})
	return shared
}

func init() {
	component.Register(
		sharedcomponent.Network,
		func() *componentCall { return &componentCall{Get()} },
		sharedcomponent.Order(62),
	)
	web.Handle("POST /call/offer", Get().OfferHandler())
	web.Handle("POST /call/answer", Get().AnswerHandler())
	web.Handle("POST /call/end", Get().EndHandler())
	detect.PauseWhen(Get().PausesWake)
	buttons.Get().Take(take)
}

func take(e buttons.Event) bool {
	if e.Kind != buttons.Tap {
		return false
	}
	if c, ok := Get().Now(); !ok || c.State != sharedcall.Ringing {
		return false
	}
	switch e.Name {
	case buttons.Action:
		Get().AnswerAudio()
	case buttons.VolumeDown:
		Get().Hangup()
	default:
		return false
	}
	return true
}
