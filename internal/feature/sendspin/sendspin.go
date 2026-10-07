package sendspin

import (
	"sync"
	"time"

	"github.com/ygelfand/echolocal/internal/android/firewall"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/timesync"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	sharedplayer "github.com/ygelfand/libcountertop/pkg/media/sendspin"
)

type Player struct{ *sharedplayer.Player }

func (p *Player) Restore(c config.Config) { p.Player.Restore(c.Sendspin.Enabled) }

var once sync.Once
var shared *Player

func Get() *Player {
	once.Do(func() {
		shared = &Player{
			sharedplayer.New(
				sharedplayer.Options{
					Output:      newOutput(),
					Arbitration: arbitration{speaker.Sound().Backgrounds()},
					Identity:    identity,
					Read:        func() bool { return config.Get().Sendspin.Enabled },
					Save:        func(v bool) error { return config.Set().Sendspin().Enabled(v) },
					Name:        func() string { return config.Get().Device.Name },
					DeviceID:    component.DevicePlayback,
					Stepped: func(f func(time.Duration)) func() {
						return timesync.Get().Synced.Listen(func(s timesync.Sync) {
							if s.Stepped() {
								f(s.After.Sub(s.Before))
							}
						})
					},
					OpenPort:  func() { firewall.Open(firewall.Sendspin, Port) },
					Advertise: advertise,
					Media:     func() sharedplayer.Media { return media.Get() },
				},
			),
		}
	})
	return shared
}
func init() { component.Register(component.Device, Get, component.Order(26)) }
