package voice

import (
	"context"

	esphome "github.com/ygelfand/go-esphome-device"
	"github.com/ygelfand/libcountertop/pkg/assistant/turn"
	"github.com/ygelfand/libcountertop/pkg/inference/wakeslots"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/activity"
	"github.com/ygelfand/echolocal/internal/feature/feedback"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/mute"
	"github.com/ygelfand/echolocal/internal/feature/recording"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/lib/wake"
)

type conversation = turn.Conversation

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
func (s voiceSound) Duck(on bool) { s.Backgrounds().Duck(on) }
func newConversation(vs *esphome.VoiceSatellite) *conversation {
	return turn.NewConversation(vs, turn.Dependencies{
		Source:       mic.Get(),
		Speaker:      speaker.Get(),
		Sound:        voiceSound{speaker.Sound()},
		Player:       media.Get(),
		Log:          activity.Get(),
		Recorder:     recording.Get(),
		Models:       func() []wake.Model { return wake.Lib().Ours() },
		HardwareTail: speaker.HardwareTail,
		Failure:      feedback.Failure,
		Cancelled:    feedback.Cancelled,
		Words:        turn.Words{Settings: config.WakeSection, Speaker: speaker.Sound()},
		Muted:        mute.Get().Muted,
		Presentation: &turn.Presentation{
			Settings: config.WakeSection,
			Shown:    &Shown,
			Effects:  &turnEffects{},
		},
	})
}

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
