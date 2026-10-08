package config

import (
	"github.com/ygelfand/libcountertop/pkg/settings/schema"
)

type Wake = schema.Wake
type Stop = schema.Stop
type WakeWord = schema.WakeWord

const (
	StopOff              = schema.StopOff
	DefaultStopThreshold = schema.DefaultStopThreshold
	DefaultThreshold     = schema.DefaultThreshold
	DefaultEffect        = schema.DefaultEffect
	DefaultTone          = schema.DefaultTone
	DefaultDelivery      = schema.DefaultDelivery
	DefaultMaxListen     = schema.DefaultMaxListen
	DefaultMaxThink      = schema.DefaultMaxThink
	DefaultFollowUp      = schema.DefaultFollowUp
	DefaultBuffer        = schema.DefaultBuffer
)

var defaultStop = schema.DefaultStop
var DefaultWakeWord = schema.DefaultWakeWord

type StopWriter struct{ st *Store }

func (w StopWriter) Threshold(v float64) error {
	return w.st.Update(func(c *Config) { c.Wake.Stop.Threshold = v })
}

type WakeWriter struct {
	st   *Store
	slot int
}

func (w WakeWriter) ID(v string) error {
	return w.word(func(word *WakeWord) { word.ID = v })
}

func (w WakeWriter) Threshold(v float64) error {
	return w.word(func(word *WakeWord) { word.Threshold = v })
}

func (w WakeWriter) Tone(v Tone) error {
	return w.word(func(word *WakeWord) { word.Tone = v })
}

func (w WakeWriter) Effect(v string) error {
	return w.word(func(word *WakeWord) { word.Effect = v })
}

func (w WakeWriter) ThinkingEffect(v string) error {
	return w.word(func(word *WakeWord) { word.ThinkingEffect = v })
}

func (w WakeWriter) ReplyingEffect(v string) error {
	return w.word(func(word *WakeWord) { word.ReplyingEffect = v })
}

func (w WakeWriter) Delivery(v Delivery) error {
	return w.word(func(word *WakeWord) { word.Delivery = v })
}

func (w WakeWriter) FollowUp(seconds int) error {
	return w.word(func(word *WakeWord) { word.FollowUp = seconds })
}

func (w WakeWriter) Buffer(ms int) error {
	return w.word(func(word *WakeWord) { word.Buffer = ms })
}

func (w WakeWriter) MaxListen(seconds int) error {
	return w.word(func(word *WakeWord) { word.MaxListen = seconds })
}

func (w WakeWriter) MaxThink(seconds int) error {
	return w.word(func(word *WakeWord) { word.MaxThink = seconds })
}

func (w WakeWriter) Recordings(count int) error {
	return w.word(func(word *WakeWord) { word.Recordings = count })
}

func (w WakeWriter) word(f func(*WakeWord)) error {
	if w.slot < 0 {
		return errSlot(w.slot)
	}
	return w.st.Update(func(c *Config) {
		for len(c.Wake.Words) <= w.slot {
			c.Wake.Words = append(c.Wake.Words, DefaultWakeWord())
		}
		f(&c.Wake.Words[w.slot])
	})
}

type Delivery = schema.Delivery

const (
	DeliveryWhole  = schema.DeliveryWhole
	DeliveryStream = schema.DeliveryStream
)

var Deliveries = schema.Deliveries

type Tone = schema.Chime

const (
	ToneNone       = schema.ChimeNone
	ToneChirp      = schema.ChimeChirp
	ToneDing       = schema.ChimeDing
	ToneRise       = schema.ChimeRise
	ToneMuteSwitch = schema.ChimeMuteSwitch
)
