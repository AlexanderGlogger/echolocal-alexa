package microphone

import (
	"log/slog"
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/say"
	setting "github.com/ygelfand/libcountertop/pkg/settings"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
)

func init() {
	component.Register(sharedcomponent.Device, Get, sharedcomponent.Order(5))
}

const (
	GainMost         = 59
	SensitivityLeast = 4
	SensitivityMost  = 20
)

const Group setting.Group = "Microphone"

type Setting = setting.Setting[config.Microphone]

type Microphone struct {
	knobs *setting.Controls[config.Microphone]
}

var (
	once   sync.Once
	shared *Microphone
)

func Get() *Microphone {
	once.Do(func() {
		shared = &Microphone{knobs: &setting.Controls[config.Microphone]{
			Table:  Table(mic.Mixings()),
			Device: component.DeviceMicrophone,
			Read:   func() config.Microphone { return config.Get().Microphone },
			Save:   store,
		}}
		component.Settings.Add(shared.knobs)
	})
	return shared
}

func Table(mixings []config.Mixing) *setting.Table[config.Microphone] {
	return setting.NewTable("microphone", []setting.Group{Group}, []Setting{
		setting.Pick(
			Setting{Name: "mixing", Group: Group, Icon: "mdi:microphone-settings"},
			mixings,
			func(m *config.Microphone) *config.Mixing { return &m.Mixing },
		),
		setting.Count(
			Setting{
				Name:  "gain",
				Group: Group,
				Icon:  "mdi:volume-plus",
				Max:   GainMost,
				Unit:  "dB",
			},
			func(m *config.Microphone) *int { return &m.Gain },
		),
		setting.Flag(
			Setting{Name: "leveling", Group: Group, Icon: "mdi:signal-variant"},
			func(m *config.Microphone) *bool { return &m.Leveling },
		),
		setting.Flag(
			Setting{
				Name:  "cancel",
				Group: Group,
				ID:    "microphone_cancel_echo",
				Icon:  "mdi:waveform",
			},
			func(m *config.Microphone) *bool { return &m.Cancel },
		),
		setting.Count(
			Setting{
				Name:  "sensitivity",
				Group: Group,
				Icon:  "mdi:motion-sensor",
				Min:   SensitivityLeast,
				Max:   SensitivityMost,
				Unit:  "dB",
			},
			func(m *config.Microphone) *int { return &m.Sensitivity },
		),
	}, setting.Messages{Text: say.T, Missing: say.Missing})
}

var applies = map[string]func(*config.Microphone) error{
	"mixing": func(m *config.Microphone) error {
		m.Mixing = mic.Get().SetMixing(m.Mixing)
		return nil
	},
	"gain": func(m *config.Microphone) error { return mic.SetGain(m.Gain) },
	"leveling": func(m *config.Microphone) error {
		mic.Get().SetLeveling(m.Leveling)
		return nil
	},
	"cancel": func(m *config.Microphone) error {
		mic.Get().SetCancelling(m.Cancel)
		return nil
	},
	"sensitivity": func(m *config.Microphone) error {
		mic.Get().SetSensitivity(m.Sensitivity)
		return nil
	},
}

func store(s Setting, v string) error {
	at := config.Get().Microphone
	if err := s.Write(&at, v); err != nil {
		return err
	}
	if err := applies[s.Name](&at); err != nil {
		return err
	}
	return setting.Store(config.MicrophoneSection, s, s.Read(&at))
}

func (m *Microphone) Name() string { return "microphone settings" }

func (m *Microphone) Entities() []esphome.Entity { return m.knobs.Entities() }

func (m *Microphone) Restore(c config.Config) {
	source := mic.Get()
	at := c.Microphone
	at.Mixing = source.SetMixing(at.Mixing)
	source.SetLeveling(at.Leveling)
	source.SetCancelling(at.Cancel)
	source.SetSensitivity(at.Sensitivity)
	m.knobs.PublishFrom(at)
	slog.Info("restored", "what", "microphone", "using", at, "asked", c.Microphone.Mixing)
}
