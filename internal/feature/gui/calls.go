//go:build board_checkers || board_cronos

package gui

import (
	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	"github.com/ygelfand/libcountertop/pkg/display/callview"
	"github.com/ygelfand/libcountertop/pkg/display/screens"
	"github.com/ygelfand/libcountertop/pkg/display/widgets"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/call"
	"github.com/ygelfand/echolocal/internal/feature/discovery"
	"github.com/ygelfand/echolocal/internal/feature/volume"
)

func init() {
	extensions = append(extensions, func(look *widgets.UI, o *screens.Options) {
		calls := callview.New(callview.Options{
			UI:    look,
			Calls: call.Get(),
			Peers: discovery.Get(),
			Voice: sharedvolume.For(volume.Get(), config.StreamVoice),
			Video: component.Board().CameraWidth > 0,
		})
		o.Boards = map[string]screens.Board{
			discovery.TabKind: {Glyph: callview.Glyph, Draw: calls.Board},
		}
		o.Views = append(o.Views, calls.Screen)
	})
}
