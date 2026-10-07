package voice

import (
	"context"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/activity"
	"github.com/ygelfand/echolocal/internal/feature/feedback"
	"github.com/ygelfand/echolocal/internal/feature/light"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/mute"
	"github.com/ygelfand/echolocal/internal/feature/recording"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/hardware/led"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/lib/wake"
	esphome "github.com/ygelfand/go-esphome-device"
	"github.com/ygelfand/libcountertop/pkg/assistant/turn"
	"github.com/ygelfand/libcountertop/pkg/inference/wakeslots"
	"time"
)

type conversation = turn.Conversation
type phase = turn.Phase

const (
	phaseIdle      = turn.Idle
	phaseListening = turn.Listening
	phaseThinking  = turn.Thinking
	phaseReplying  = turn.Replying
)

type voiceSound struct{ *speaker.Driver }

func (s voiceSound) Claim(name string, fn func(context.Context, turn.Speaker) error) turn.Claim {
	return s.Driver.Claim(
		name,
		func(ctx context.Context, p *speaker.Player) error { return fn(ctx, p) },
	)
}
func newConversation(vs *esphome.VoiceSatellite) *conversation {
	var effect *led.Claim
	return turn.NewConversation(vs, turn.Options{
		Source:  mic.Get(),
		Speaker: speaker.Get(),
		Sound: voiceSound{
			speaker.Sound(),
		},
		Player:       media.Get(),
		Log:          activity.Get(),
		Recorder:     recording.Get(),
		Models:       func() []wake.Model { return wake.Lib().Ours() },
		Adapting:     mic.Get().SetAdapting,
		Duck:         speaker.Sound().Backgrounds().Duck,
		HardwareTail: speaker.HardwareTail,
		Shown:        Shown.Emit,
		Failure:      feedback.Failure,
		Words: turn.Words{
			MaxListen:   wakeword.MaxListen,
			MaxThink:    wakeword.MaxThink,
			FollowUp:    wakeword.FollowUp,
			Buffer:      wakeword.Buffer,
			ChimeLength: wakeword.ChimeLength,
			Tones:       wakeword.Tones,
			Chime:       wakeword.Chime,
			Delivery:    func(slot int) string { return string(wakeword.Delivery(slot)) },
		},
		Muted:     mute.Get().Muted,
		Cancelled: feedback.Cancelled,
		Begin:     func() { effect = led.Get().Claim(led.PriorityTurn) },
		End:       func() { effect.Release() },
		Elsewhere: func() { led.Get().Busy().Flash(led.WorkElsewhere, 900*time.Millisecond) },
		Effect: func(slot int, phase turn.Phase) {
			if phase == turn.Idle {
				effect.Clear()
				return
			}
			base := light.Get().Base()
			if phase == turn.Listening {
				if name := wakeword.Effect(slot); name != "" {
					effect.Play(name, base)
				}
				return
			}
			name := wakeword.ThinkingEffect(slot)
			if phase == turn.Replying {
				name = wakeword.ReplyingEffect(slot)
			}
			switch name {
			case "":
				if fallback := wakeword.Effect(slot); fallback != "" {
					effect.PlayReversed(fallback, base)
				}
			case component.EffectNone:
				effect.Clear()
			default:
				effect.Play(name, base)
			}
		},
	})
}

// wanted retains assistant positions and explicit disables. Only a fresh configuration gets a default.
func wanted(models []wake.Model, count int) []string {
	saved := config.Get().Wake.Words
	var ids []string
	if saved != nil {
		ids = make([]string, len(saved))
		for i, word := range saved {
			ids[i] = word.ID
		}
	}
	installed := make([]string, len(models))
	for i, m := range models {
		installed[i] = m.ID
	}
	return wakeslots.Initial(ids, installed, wake.DefaultModel, count)
}
