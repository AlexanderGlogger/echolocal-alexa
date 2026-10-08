package sendspin

import (
	"log/slog"
	"sync"

	sharedplayer "github.com/ygelfand/libcountertop/pkg/media/sendspin"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/firewall"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/timesync"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
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
		OpenPort: func() {
			if err := firewall.Open(firewall.Sendspin, Port); err != nil {
				slog.Error("opening the sendspin port failed", "port", Port, "err", err)
			}
		},
	})
})

func Get() *sharedplayer.Player { return get() }
func init() {
	component.Register(
		sharedcomponent.Device,
		func() *componentPlayer { return &componentPlayer{Get()} },
		sharedcomponent.Order(26),
	)
}
