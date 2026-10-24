//go:build !board_biscuit

package playout

import (
	"context"

	"github.com/ygelfand/libcountertop/pkg/display/geometry"
	"github.com/ygelfand/libcountertop/pkg/display/playoutscreen"
	"github.com/ygelfand/libcountertop/pkg/display/video"
	"github.com/ygelfand/libcountertop/pkg/media/playout"
	"github.com/ygelfand/libcountertop/pkg/media/videostream"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/feature/videoplayer"
	"github.com/ygelfand/echolocal/internal/hardware/display"
)

func screen() playout.Screen {
	return playoutscreen.New(playoutscreen.Options{
		NewPage: videoplayer.NewPage,
		NewUnder: func() *videostream.Beneath {
			return &videostream.Beneath{
				Rotation: func() geometry.Orientation { return display.Get().Orientation() },
				Size:     func() (int, int) { return display.Get().Native() },
			}
		},
		Video: func(ctx context.Context, b *videostream.Beneath, s videostream.Stream) (videostream.Report, error) {
			player := video.Player{
				Helper:         display.Get().Helper(),
				SecureDecoders: component.Board().SecureDecoders,
			}
			return player.On(ctx, b, s)
		},
	})
}
