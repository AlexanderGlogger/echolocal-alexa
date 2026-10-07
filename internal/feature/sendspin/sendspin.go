package sendspin

import (
	"sync"

	"github.com/ygelfand/echolocal/internal/android/firewall"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/timesync"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	sharedplayer "github.com/ygelfand/libcountertop/pkg/media/sendspin"
)

type componentPlayer struct{ *sharedplayer.Player }

func (p *componentPlayer) Restore(c config.Config) { p.Player.Restore(c.Sendspin.Enabled) }

var get = sync.OnceValue(func() *sharedplayer.Player {
	return sharedplayer.New(sharedplayer.Dependencies{
		Output:      newOutput(),
		Arbitration: speaker.Sound().Backgrounds(),
		Device:      identity{},
		Settings:    config.SendspinSection,
		DeviceID:    component.DevicePlayback,
		Clock:       timesync.Get(),
		Media:       media.Get(),
		OpenPort:    func() { firewall.Open(firewall.Sendspin, Port) },
	})
})

func Get() *sharedplayer.Player { return get() }
func init() {
	component.Register(
		component.Device,
		func() *componentPlayer { return &componentPlayer{Get()} },
		component.Order(26),
	)
}
