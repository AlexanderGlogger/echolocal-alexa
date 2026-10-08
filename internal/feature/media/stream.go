package media

import (
	"github.com/ygelfand/libcountertop/pkg/media/pcm"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func NewStream(sound *speaker.Driver, out *speaker.Player, changed func()) *pcm.Stream {
	return pcm.New(pcm.Options{
		Sink:        out,
		Rate:        speaker.Rate,
		Channels:    speaker.Channels,
		Changed:     changed,
		DuckDB:      func() float64 { return config.Get().Media.DuckDB },
		Arbitration: sound.Backgrounds(),
	})
}
