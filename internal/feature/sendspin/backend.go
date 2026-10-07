package sendspin

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/layout"
	core "github.com/ygelfand/libcountertop/pkg/audio/sendspin"
	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	sharedplayer "github.com/ygelfand/libcountertop/pkg/media/sendspin"
)

const Port = core.Port

type outputSink struct{ p *speaker.Player }

func (s outputSink) Attach(r core.Renderer) { s.p.Attach(r) }
func (s outputSink) Written() uint64        { return s.p.Written() }

func newOutput() *core.Output {
	return core.NewOutput(outputSink{speaker.Get()}, core.Controls{
		Volume: sharedvolume.Steps(media.Get(), speaker.VolumeSteps), Media: media.Get(), Settings: config.MediaSection, HardwareTail: speaker.HardwareTail,
	})
}

type identity struct{}

func (identity) Name() string { return config.Get().Device.Name }
func (identity) Identity() core.Identity {
	mac, _ := layout.FactoryMAC()
	model := config.Get().Device.Board.Model
	return core.Identity{
		ID:           mac,
		Model:        model,
		Manufacturer: layout.Manufacturer,
		Version:      layout.Version,
		ArtworkSize:  sharedplayer.ArtworkSize,
	}
}
