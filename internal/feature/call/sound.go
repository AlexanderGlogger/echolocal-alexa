//go:build board_checkers || board_cronos

package call

import (
	"context"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

type device struct{}

func (device) Frames() (<-chan []int16, func()) { return mic.Get().Listen("call") }

func (device) Play(pcm []int16) {
	speaker.Get().PlayStream(config.StreamVoice, pcm)
}

func (device) Queued() int { return speaker.Get().Queued() }

func (device) Claim(name string, run func(context.Context) error) {
	speaker.Sound().
		Claim(name, func(ctx context.Context, _ *speaker.Player) error { return run(ctx) })
}
func (device) Sounding()    {}
func (device) Duck(on bool) { speaker.Sound().Backgrounds().Duck(on) }
func (device) Chime(level float64, notes ...speaker.Note) {
	speaker.Sound().Interject(func(p *speaker.Player) { p.Chime(level, notes...) })
}
