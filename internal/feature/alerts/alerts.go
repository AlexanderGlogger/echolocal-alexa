package alerts

import (
	"sync"

	shared "github.com/ygelfand/libcountertop/pkg/audio/alerts"
	"github.com/ygelfand/libcountertop/pkg/audio/tone"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func init() { component.Register(sharedcomponent.Device, Get, sharedcomponent.Order(30)) }

var get = sync.OnceValue(func() *shared.Alerts {
	a := shared.New(config.AlertsSection, component.DevicePlayback, func(notes []tone.Note) {
		speaker.Sound().
			Interject(func(p *speaker.Player) { p.Chime(config.StreamAlerts, shared.Level, notes...) })
	})
	component.Settings.Add(a.Controls)
	return a
})

func Get() *shared.Alerts { return get() }
