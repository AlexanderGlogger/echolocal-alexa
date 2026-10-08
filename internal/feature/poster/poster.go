package poster

import (
	"sync"

	sharedposter "github.com/ygelfand/libcountertop/pkg/display/poster"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/metrics"
	"github.com/ygelfand/echolocal/internal/hardware/touch"
)

type componentPoster struct{ *sharedposter.Poster }

func (p *componentPoster) Restore(c config.Config) { p.Poster.Restore(c.Poster) }

var get = sync.OnceValue(func() *sharedposter.Poster {
	return sharedposter.New(
		sharedposter.Dependencies{
			Settings: config.PosterSection,
			Touch:    touch.Get(),
			Ambient:  ambient,
			DeviceID: component.DeviceScreen,
		},
	)
})

func Get() *sharedposter.Poster { return get() }
func init() {
	component.Register(
		component.Device,
		func() *componentPoster { return &componentPoster{Get()} },
		component.Order(25),
		component.Needs(board.Panel),
	)
}
func ambient() (float64, bool) { r := metrics.Reader{}.Lux(""); return r.Value, r.Known }
