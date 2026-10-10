//go:build board_biscuit

package chromecast

import (
	"github.com/ygelfand/libcountertop/pkg/media/cast"
	"github.com/ygelfand/libcountertop/pkg/media/cast/protocols/unsupported"
	"github.com/ygelfand/libcountertop/pkg/media/cast/protocols/youtube"
	"github.com/ygelfand/libcountertop/pkg/media/castoutput"
	"github.com/ygelfand/libcountertop/pkg/media/playback"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
)

func needs() []sharedcomponent.Option { return nil }

func apps() []cast.Maker { return []cast.Maker{unsupported.Maker, youtube.Maker(nil)} }

func pictures() func() playback.Target { return nil }

func logo() func() []byte { return nil }

func screen() castoutput.Screen { return nil }
