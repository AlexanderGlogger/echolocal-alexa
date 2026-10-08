package wakeword

import (
	"time"

	sharedtone "github.com/ygelfand/libcountertop/pkg/audio/tone"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func saved(slot int) config.WakeWord { return config.Get().Wake.Slot(slot) }

func Threshold(slot int) float64 { return saved(slot).Threshold }

func Chime(slot int) { speaker.Sound().Chime(sharedtone.Wake(saved(slot).Tone)) }

func ThinkingEffect(slot int) string { return saved(slot).ThinkingEffect }
func ReplyingEffect(slot int) string { return saved(slot).ReplyingEffect }

func Tones(slot int) bool { return saved(slot).Tone != config.ToneNone }

func ChimeLength(slot int) time.Duration {
	return sharedtone.Length(sharedtone.Wake(saved(slot).Tone))
}

func Delivery(slot int) config.Delivery { return saved(slot).Delivery }

func Buffer(slot int) time.Duration {
	return time.Duration(saved(slot).Buffer) * time.Millisecond
}

func FollowUp(slot int) time.Duration {
	return time.Duration(saved(slot).FollowUp) * time.Second
}

func MaxListen(slot int) time.Duration {
	return time.Duration(saved(slot).MaxListen) * time.Second
}

func MaxThink(slot int) time.Duration {
	return time.Duration(saved(slot).MaxThink) * time.Second
}

func Effect(slot int) string {
	if e := saved(slot).Effect; e != component.EffectNone {
		return e
	}
	return ""
}
