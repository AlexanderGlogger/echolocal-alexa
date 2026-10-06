package sendspin

import (
	"github.com/Sendspin/sendspin-go/pkg/protocol"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/layout"
	core "github.com/ygelfand/libcountertop/pkg/audio/sendspin"
)

const Port = core.Port

type outputSink struct{ p *speaker.Player }

func (s outputSink) Attach(r core.Renderer) { s.p.Attach(r) }
func (s outputSink) Written() uint64        { return s.p.Written() }

type arbitration struct{ a *speaker.Arbiter }

func (a arbitration) Took(p core.Producer) { a.a.Took(p) }
func (a arbitration) Gave(p core.Producer) { a.a.Gave(p) }
func newOutput() *core.Output {
	return core.NewOutput(outputSink{speaker.Get()}, core.Controls{
		Volume:       func() int { return config.Get().Speaker.Volume * 100 / speaker.VolumeSteps },
		SetVolume:    func(level int) { media.Get().Set(level * speaker.VolumeSteps / 100) },
		SetMuted:     func(on bool) { media.Get().Mute(on) },
		DuckDB:       func() float64 { return float64(config.Get().Media.DuckDB) },
		HardwareTail: speaker.HardwareTail,
	})
}
func newListener(p *Player) *core.Listener {
	return core.NewListener(
		p.out,
		arbitration{speaker.Sound().Backgrounds()},
		func() core.Identity {
			mac, _ := layout.FactoryMAC()
			model := config.Get().Device.Board.Model
			return core.Identity{
				ID:           mac,
				Model:        model,
				Manufacturer: layout.Manufacturer,
				Version:      layout.Version,
				ArtworkSize:  artworkSize,
			}
		},
		core.Callbacks{
			State:      p.setState,
			Connection: p.holds,
			Began:      func() { media.Get().Began(p) },
			Grouped:    p.grouped,
			Artwork:    p.drew,
			Metadata:   func(m *protocol.MetadataState, now func() int64) { p.plays(p.track().merge(m)) },
		},
	)
}
