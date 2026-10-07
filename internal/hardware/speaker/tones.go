package speaker

import (
	"github.com/ygelfand/echolocal/internal/config"
	sharedtone "github.com/ygelfand/libcountertop/pkg/audio/tone"
)

// The sounds the device makes about itself, as opposed to anything it was asked to play.
//
// Direction carries the meaning — rising for on, falling for off — and a hold is two notes where a
// press is one, so they are told apart without looking.
const toneLevel = sharedtone.Level

// WakeTones lists them in the order they are offered.
func WakeTones() []config.Tone {
	return []config.Tone{config.ToneNone, config.ToneChirp, config.ToneDing, config.ToneRise}
}

// MuteTones is what the mute switch can sound like, in the order they are offered.
func MuteTones() []config.Tone {
	return []config.Tone{
		config.ToneNone, config.ToneMuteSwitch, config.ToneChirp, config.ToneDing, config.ToneRise,
	}
}

// MuteTone is what t sounds like for the mute switch.
func MuteTone(t config.Tone, muted bool) []Note {
	if t == config.ToneMuteSwitch {
		if muted {
			return sharedtone.ToneMute
		}
		return sharedtone.ToneUnmute
	}
	return sharedtone.Wake(t)
}

// Chime plays a tone alongside whatever is playing rather than instead of it: pressing volume during
// a reply should beep and leave the reply alone.
func (d *Driver) Chime(notes []Note) {
	if d == nil || len(notes) == 0 {
		return
	}
	d.Interject(func(p *Player) { p.Chime(toneLevel, notes...) })
}
