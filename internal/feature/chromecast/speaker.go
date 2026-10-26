//go:build board_biscuit

package chromecast

import (
	"github.com/ygelfand/libcountertop/pkg/media/cast"
	"github.com/ygelfand/libcountertop/pkg/media/cast/protocols/youtube"
	"github.com/ygelfand/libcountertop/pkg/media/castreceiver"
	"github.com/ygelfand/libcountertop/pkg/media/playback"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
)

func needs() []sharedcomponent.Option { return nil }

func configureSettings(r *castreceiver.Receiver) {
	r.Settings().Table = r.Settings().Table.Without("youtube.skip")
}

func apps() []cast.Maker { return []cast.Maker{youtube.Maker} }

func pictures() func() playback.Target { return nil }

func notice() func(title, body string) { return nil }

func logo() func() []byte { return nil }
