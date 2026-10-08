package speaker

import (
	sharedtone "github.com/ygelfand/libcountertop/pkg/audio/tone"

	"github.com/ygelfand/echolocal/internal/config"
)

const toneLevel = sharedtone.Level

func WakeTones() []config.Tone {
	return []config.Tone{config.ToneNone, config.ToneChirp, config.ToneDing, config.ToneRise}
}

func MuteTones() []config.Tone {
	return []config.Tone{
		config.ToneNone, config.ToneMuteSwitch, config.ToneChirp, config.ToneDing, config.ToneRise,
	}
}

func MuteTone(t config.Tone, muted bool) []Note {
	if t == config.ToneMuteSwitch {
		if muted {
			return sharedtone.ToneMute
		}
		return sharedtone.ToneUnmute
	}
	return sharedtone.Wake(t)
}

func (d *Driver) Chime(notes []Note) {
	if d == nil || len(notes) == 0 {
		return
	}
	d.Interject(func(p *Player) { p.Chime(toneLevel, notes...) })
}
