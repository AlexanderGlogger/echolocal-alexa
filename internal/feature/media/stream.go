package media

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/libcountertop/pkg/media/pcm"
)

type Stream = pcm.Stream

func NewStream(sound *speaker.Driver, out *speaker.Player, changed func()) *Stream {
	arb := sound.Backgrounds()
	return pcm.New(
		pcm.Options{Sink: out, Rate: speaker.Rate, Channels: speaker.Channels, Changed: changed,
			DuckDB:      func() float64 { return config.Get().Media.DuckDB },
			Arbitration: arb,
		},
	)
}
