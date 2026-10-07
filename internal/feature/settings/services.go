//go:build board_checkers || board_cronos

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/chromecast"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	sharedcast "github.com/ygelfand/libcountertop/pkg/display/settings/castpages"
)

func servicePages() *sharedcast.Pages {
	return sharedcast.New(sharedcast.Options{Read: func() sharedcast.State {
		c := config.Get()
		return sharedcast.State{
			Receiver:       c.Cast.Receiver,
			OnDemand:       c.Cast.YouTube.OnDemand,
			Skip:           c.Cast.YouTube.Skip,
			LiveDelay:      c.Cast.YouTube.LiveDelay,
			PrimePersist:   c.Cast.Prime.Persist,
			PrimeSkipIntro: c.Cast.Prime.SkipIntro,
		}
	}, Push: shell.Get().Push, SetReceiver: chromecast.Get().SetReceiver, SetSkip: chromecast.Get().SetSkip, SetLoungeOnDemand: chromecast.Get().SetLoungeOnDemand, SetLiveDelay: chromecast.Get().SetLiveDelay, PrimeRegistered: chromecast.Get().PrimeRegistered, SetPrimePersist: chromecast.Get().SetPrimePersist, SetPrimeSkipIntro: chromecast.Get().SetPrimeSkipIntro, ResetPrime: chromecast.Get().ResetPrime})
}
