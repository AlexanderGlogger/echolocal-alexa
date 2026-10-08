package chromecast

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/ygelfand/libcountertop/pkg/audio/resample"
	"github.com/ygelfand/libcountertop/pkg/display/surface"
	"github.com/ygelfand/libcountertop/pkg/display/video"
	"github.com/ygelfand/libcountertop/pkg/media/cast/preferences"
	"github.com/ygelfand/libcountertop/pkg/media/castreceiver"
	"github.com/ygelfand/libcountertop/pkg/media/playback"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dhcp"
	"github.com/ygelfand/echolocal/internal/feature/timesync"
	"github.com/ygelfand/echolocal/internal/feature/volume"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/ui"
)

func init() {
	preferences.Configure(castreceiver.Preferences("echolocal", config.CastSection))
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(60),
		component.Needs(board.Panel),
	)
}

var get = sync.OnceValue(build)

func Get() *castreceiver.Receiver { return get() }

func build() *castreceiver.Receiver {
	return castreceiver.NewReceiver(castreceiver.ReceiverOptions{
		Settings: config.CastSection,
		DeviceID: component.DevicePlayback,
		Network:  func(ctx context.Context) bool { return dhcp.Get().Wait(ctx) },
		Server: castreceiver.Config{
			Name:            deviceName,
			Model:           component.Board().Model,
			Version:         layout.Version,
			Hardware:        hardware,
			Host:            layout.Slug,
			Logo:            ui.Logo().PNG,
			Address:         func() string { return config.Get().Network.Address },
			Timezone:        func() string { return timesync.Get().Zone() },
			CredentialsPath: layout.CastCredentialsPath,
			AuthorityPath:   layout.CastAuthorityPath,
			AppDir:          layout.CastAppDir,
			Synced: func(do func()) func() {
				return timesync.Get().Synced.Listen(func(timesync.Sync) { do() })
			},
			Volume: mediaVolume{},
			Video: func() playback.Target {
				return video.Target(display.Get(), component.Board().MaxFPS)
			},
			Output: func(ended func()) playback.Output {
				o := newOutput()
				o.SetEnded(ended)
				return o
			},
			Resample: func(from int) func([]int16) []int16 {
				return resample.NewRational(from, speaker.Rate, speaker.Channels).Run
			},
			Surface: func() *surface.Client { return display.Get().Helper() },
		},
	})
}

type mediaVolume struct{}

func (mediaVolume) Level() int { return volume.Get().Level(config.StreamMain) }

func (mediaVolume) SetLevel(level int) { volume.Get().Set(config.StreamMain, level) }

func (mediaVolume) Listen(changed func(int)) func() {
	return volume.Get().Changed.Listen(func(c volume.Change) {
		if c.Stream == config.StreamMain {
			changed(c.Level)
		}
	})
}

func deviceName() string {
	if name := config.Get().Device.Name; name != "" {
		return name
	}
	return component.Board().DefaultName
}

func hardware() string {
	raw, err := os.ReadFile(layout.MACPath)
	if err != nil {
		slog.Warn("the cast identity has no hardware address to come from", "err", err)
		return ""
	}
	return layout.MAC(string(raw))
}
