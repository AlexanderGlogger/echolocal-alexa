package timer

import (
	"math"
	"slices"
	"sync"
	"time"

	"github.com/ygelfand/libcountertop/pkg/assistant/timer"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/feature/alerts"
	"github.com/ygelfand/echolocal/internal/hardware/led"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func init() {
	component.Register(sharedcomponent.Device, Get, sharedcomponent.Order(31))
}

var (
	countdownColor = led.Color{R: 0xFF, G: 0x8C, B: 0x00}
	alarmColor     = led.Color{R: 0xFF, G: 0x40, B: 0x00}
)

var get = sync.OnceValue(func() *timer.Timers {
	countdown := led.Get().Claim(led.PriorityTimer)
	alarm := led.Get().Claim(led.PriorityAlarm)
	return timer.New(timer.Options{
		Alarm: func() func() {
			sound := speaker.Sound()
			sound.Backgrounds().Duck(true)
			alarm.Play(led.EffectPulse, alarmColor)
			return func() {
				alarm.Clear()
				sound.Backgrounds().Duck(false)
			}
		},
		Chime:     alerts.Get().RingTimer,
		RingFor:   alerts.Get().RingFor,
		Countdown: &ring{claim: countdown},
	})
})

func Get() *timer.Timers { return get() }

type ring struct {
	claim *led.Claim
	shown []led.Color
}

func (r *ring) Show(left, total time.Duration) { r.paint(frame(left, total)) }

func (r *ring) Clear() { r.paint(nil) }

func (r *ring) paint(next []led.Color) {
	if !r.claim.Present() || slices.Equal(r.shown, next) {
		return
	}
	r.shown = next

	if next == nil {
		r.claim.Clear()
		return
	}
	r.claim.Paint(next)
}

func frame(left, total time.Duration) []led.Color {
	lit := float64(led.Segments) * left.Seconds() / math.Max(total.Seconds(), 1)

	out := make([]led.Color, led.Segments)
	for i := range out {
		switch part := lit - float64(i); {
		case part >= 1:
			out[i] = countdownColor
		case part > 0:
			out[i] = dim(countdownColor, part)
		}
	}
	return out
}

func dim(c led.Color, by float64) led.Color {
	scale := func(v byte) byte { return byte(math.Round(float64(v) * by)) }
	return led.Color{R: scale(c.R), G: scale(c.G), B: scale(c.B)}
}
