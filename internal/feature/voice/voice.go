package voice

import (
	"context"
	"log/slog"
	"sync"

	"github.com/ygelfand/libcountertop/pkg/inference/wakeslots"

	esphome "github.com/ygelfand/go-esphome-device"
	"google.golang.org/protobuf/proto"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/timer"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/lib/wake"
)

func init() {
	component.Register(component.Device, Get, component.Order(30))
}

const Features = esphome.DefaultVoiceFeatures |
	esphome.FeatureSpeaker |
	esphome.FeatureAnnounce |
	esphome.FeatureStartConversation |
	esphome.FeatureTimers

type Voice struct {
	vs        *esphome.VoiceSatellite
	turn      *conversation
	selection wakeslots.Selection
}

var (
	once   sync.Once
	shared *Voice
)

func Get() *Voice {
	once.Do(func() { shared = build() })
	return shared
}

func build() *Voice {
	ours := wake.Lib().Ours()
	slots := wanted(ours, wakeword.Slots)

	v := &Voice{
		vs: &esphome.VoiceSatellite{
			ActiveWakeWords:     chosen(slots),
			MaxActiveWakeWords:  wakeword.Slots,
			OnExternalWakeWords: wakeword.Answer,
		},
	}
	v.selection.Set(slots)
	v.turn = newConversation(v.vs)
	v.turn.Word = v.selection.ID
	slog.Info("wake words", "ours", len(ours), "slots", slots)

	v.vs.OnTimer = timer.Get().Event
	v.vs.OnSubscribed = func(subscribed bool) {
		if !subscribed {
			timer.Get().Forget()
		}
	}

	wakeword.Requested.Listen(v.Start)

	buttons.Get().Events.Listen(func(e buttons.Event) {
		if e.Name != buttons.Action {
			return
		}
		switch e.Kind {
		case buttons.Tap:
			v.Action()
		case buttons.Hold:
			v.ActionHold()
		}
	})
	return v
}

func (v *Voice) Name() string { return "conversation" }

func (v *Voice) Handle(ctx context.Context, c *esphome.Conn, msg proto.Message) error {
	return v.vs.Handle(ctx, c, msg)
}

func (v *Voice) Run(ctx context.Context) error {
	v.turn.Run(ctx)
	return nil
}

func (v *Voice) Ready() bool { return v.vs.Subscribed() }

func (v *Voice) Start(slot int) { v.turn.Start(slot) }

func (v *Voice) Action() {
	if v.Stop() {
		return
	}

	// Home Assistant falls back to the first pipeline for a wake word with no phrase.
	v.turn.Start(0)
}

func (v *Voice) Interrupt() {
	if v.turn.Phase() == phaseListening {
		slog.Debug("stop word ignored, the turn is listening")
		return
	}
	if !speaker.Sound().Busy() && !timer.Get().Ringing() {
		if playing, _ := media.Get().Playing(); !playing {
			slog.Debug("stop word ignored, nothing to stop")
			return
		}
	}
	v.Stop()
}

func (v *Voice) Stop() bool {
	if timer.Get().Stop() {
		return true
	}

	if v.turn.Busy() {
		v.turn.Cancel()
		return true
	}

	if sound := speaker.Sound(); sound.Busy() {
		sound.Silence()
		return true
	}

	if playing, _ := media.Get().Playing(); playing {
		media.Get().Pause()
		return true
	}
	return false
}

func (v *Voice) ActionHold() { v.turn.Start(1) }

// Home Assistant takes the active echo as authoritative and reverts a slot missing from it.
func (v *Voice) OnWakeWord(load func(ids []string) []string, selected func()) {
	v.vs.OnSetActiveWakeWords = func(ids []string) {
		accepted := load(ids)
		v.SetSlots(accepted)

		for slot := range wakeword.Slots {
			id := ""
			if slot < len(accepted) {
				id = accepted[slot]
			}
			if err := config.Set().Wake(slot).ID(id); err != nil {
				slog.Error("saving the wake word failed", "slot", slot+1, "err", err)
			}
		}
		if len(chosen(accepted)) != len(chosen(ids)) {
			slog.Warn("some wake words were refused", "asked", ids, "running", accepted)
		}
		selected()
	}
}

func (v *Voice) Slots() []string { return v.selection.Get() }
func (v *Voice) SetSlots(slots []string) {
	v.selection.Set(slots)
	v.vs.ActiveWakeWords = chosen(slots)
	slog.Info("wake words listening", "slots", slots)
}
func chosen(slots []string) []string { return wakeslots.Advertised(slots) }
