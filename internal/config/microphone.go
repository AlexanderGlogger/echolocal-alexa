package config

import (
	"github.com/ygelfand/libcountertop/pkg/settings/schema"

	"github.com/ygelfand/echolocal/internal/defaults"
)

type Microphone struct {
	Muted     bool `json:"muted"`
	LEDBright bool `json:"led_bright"`

	MuteSound schema.Chime `json:"mute_sound"`

	Gain int `json:"gain"`

	Leveling bool `json:"leveling"`

	Mixing Mixing `json:"mixing"`

	Cancel bool `json:"cancel"`

	Sensitivity int `json:"sensitivity"`

	VisualizerLift int `json:"visualizerLift"`
}

const (
	DefaultMuted     = false
	DefaultLEDBright = true
	DefaultMuteSound = schema.ChimeMuteSwitch

	DefaultMixing = MixCenter

	DefaultLeveling = true

	DefaultCancel = true
)

func defaultMicrophone(d defaults.Set) Microphone {
	return Microphone{
		Muted:       DefaultMuted,
		LEDBright:   DefaultLEDBright,
		MuteSound:   DefaultMuteSound,
		Gain:        d.MicGain,
		Leveling:    DefaultLeveling,
		Mixing:      DefaultMixing,
		Cancel:      DefaultCancel,
		Sensitivity: d.Sensitivity,

		VisualizerLift: d.VisualizerLift,
	}
}

func (w MicrophoneWriter) VisualizerLift(db int) error {
	return w.st.Update(func(c *Config) { c.Microphone.VisualizerLift = db })
}

type MicrophoneWriter struct{ st *Store }

func (w MicrophoneWriter) Muted(v bool) error {
	return w.st.Update(func(c *Config) { c.Microphone.Muted = v })
}

func (w MicrophoneWriter) LEDBright(v bool) error {
	return w.st.Update(func(c *Config) { c.Microphone.LEDBright = v })
}

func (w MicrophoneWriter) MuteSound(v schema.Chime) error {
	return w.st.Update(func(c *Config) { c.Microphone.MuteSound = v })
}

func (w MicrophoneWriter) Gain(db int) error {
	return w.st.Update(func(c *Config) { c.Microphone.Gain = db })
}

func (w MicrophoneWriter) Leveling(v bool) error {
	return w.st.Update(func(c *Config) { c.Microphone.Leveling = v })
}

func (w MicrophoneWriter) Mixing(v Mixing) error {
	return w.st.Update(func(c *Config) { c.Microphone.Mixing = v })
}

type Mixing string

const (
	MixCenter Mixing = "center"

	MixDelaySum Mixing = "delay-sum"

	MixBeamformer Mixing = "beamformer"
)

func (m Mixing) Label() string {
	switch m {
	case MixCenter:
		return "Center mic"
	case MixDelaySum:
		return "Delay and sum"
	case MixBeamformer:
		return "Beamformer"
	}
	return string(m)
}
