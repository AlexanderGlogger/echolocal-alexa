package livecam

import (
	"fmt"
	"sync"

	"github.com/ygelfand/libcountertop/pkg/camera/camerafeed"
	"github.com/ygelfand/libcountertop/pkg/camera/live"
	"github.com/ygelfand/libcountertop/pkg/camera/session"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"

	"github.com/ygelfand/echolocal/internal/android/prop"
	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/mtkcamera"
	"github.com/ygelfand/echolocal/internal/hardware/privacy"
	"github.com/ygelfand/echolocal/internal/layout"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(36),
		component.Needs(board.Camera),
	)
}

type Camera struct{ *live.Camera }

func (c Camera) Restore(config.Config) { c.Publish() }

var vendor = live.Vendor{
	EVMin:  -3,
	EVMax:  3,
	Scenes: []string{"auto", "portrait", "landscape", "night"},
	Balances: []string{
		"auto",
		"incandescent",
		"fluorescent",
		"warm-fluorescent",
		"daylight",
		"cloudy-daylight",
		"twilight",
		"shade",
	},
	Effects: []string{"none", "mono", "negative", "sepia", "aqua", "whiteboard", "blackboard"},
	ISOs:    []string{"auto", "100", "200", "400", "800", "1600"},
	Levels:  true,
	Rates:   []int{15, 20, 30},
}

var Get = sync.OnceValue(func() Camera {
	b := component.Board()
	o := live.Options{
		Vendor: vendor,
		Streams: live.Streams{
			Main:        []string{"1280x720", "864x480", "640x480"},
			MainDefault: fmt.Sprintf("%dx%d", b.CameraWidth, b.CameraHeight),
		},
		StillOnKeyframe: true,
		Open:            func(c camerafeed.Config) (session.Transport, error) { return mtkcamera.Open(c) },
		Muted:           muted,
		Restart:         func() error { return prop.Restart(layout.CameraService) },
		Saved:           func() schema.Camera { return config.Get().Camera },
		Writer:          func() schema.CameraWriter { return config.Set().Camera() },
		Device:          component.DeviceCamera,
	}
	if b.NativeAPI >= 26 {
		o.StillOnKeyframe = false
		o.Turn = func() int { return 0 }
		o.Vendor.Levels = false
		o.Vendor.Sliders = true
		o.Vendor.ISOs = nil
		o.Vendor.Noises = []string{"off", "fast", "high_quality", "minimal"}
	}
	c := Camera{live.New(o)}
	component.Settings.Add(c.Controls())
	return c
})

func muted() bool {
	if !component.Board().Has(board.MicMutesCamera) {
		return false
	}
	m, err := privacy.Microphone()
	if err != nil {
		return false
	}
	cut, err := m.Get()
	return err == nil && cut
}
