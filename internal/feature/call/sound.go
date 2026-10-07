//go:build board_checkers || board_cronos

package call

import (
	"context"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	sharedcall "github.com/ygelfand/libcountertop/pkg/media/call"
	"github.com/ygelfand/libcountertop/pkg/media/rtc"
)

const maxQueued = rtc.PlayRate / 5

type device struct{}

func (device) Frames() (<-chan []int16, func()) { return mic.Get().Listen("call") }

func (device) Play(pcm []int16) {
	s := speaker.Get()
	if s.Queued() > maxQueued {
		return
	}
	s.PlayStream(config.StreamVoice, pcm)
}

func (device) Claim(name string, run func(context.Context) error) {
	speaker.Sound().Claim(name, func(ctx context.Context, _ *speaker.Player) error { return run(ctx) })
}
func (device) Sounding()    {}
func (device) Duck(on bool) { speaker.Sound().Backgrounds().Duck(on) }
func (device) Chime(level float64, notes ...sharedcall.Note) {
	tones := make([]speaker.Note, len(notes))
	for i, note := range notes {
		tones[i] = speaker.Note{Freq: note.Freq, Ms: note.Ms}
	}
	speaker.Sound().Interject(func(p *speaker.Player) { p.Chime(level, tones...) })
}
