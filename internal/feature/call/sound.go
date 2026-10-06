//go:build board_checkers || board_cronos

package call

import (
	"context"
	"time"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/lib/rtc"
)

const (
	ringEvery     = 3 * time.Second
	ringLevel     = 0.5
	ringbackEvery = 4 * time.Second
	ringbackLevel = 0.15
	maxQueued     = rtc.PlayRate / 5
)

var (
	ringTone     = []speaker.Note{{Freq: 1047, Ms: 110}, {Freq: 1319, Ms: 110}, {Freq: 1047, Ms: 110}, {Freq: 1319, Ms: 110}}
	ringbackTone = []speaker.Note{{Freq: 440, Ms: 400}, {Ms: 200}, {Freq: 440, Ms: 400}}
)

func ring(ctx context.Context) bool {
	sound := speaker.Sound()
	sound.Backgrounds().Duck(true)
	defer sound.Backgrounds().Duck(false)
	over := time.After(ringFor)
	for {
		sound.Interject(func(p *speaker.Player) { p.Chime(ringLevel, ringTone...) })
		select {
		case <-ctx.Done():
			return false
		case <-over:
			return true
		case <-time.After(ringEvery):
		}
	}
}

func ringback(ctx context.Context) {
	sound := speaker.Sound()
	for {
		sound.Interject(func(p *speaker.Player) { p.Chime(ringbackLevel, ringbackTone...) })
		select {
		case <-ctx.Done():
			return
		case <-time.After(ringbackEvery):
		}
	}
}

type device struct{}

func (device) Frames() (<-chan []int16, func()) { return mic.Get().Listen("call") }

func (device) Play(pcm []int16) {
	s := speaker.Get()
	if s.Queued() > maxQueued {
		return
	}
	s.PlayStream(config.StreamVoice, pcm)
}
