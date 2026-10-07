//go:build board_checkers || board_cronos || board_rook

package gui

import (
	"github.com/ygelfand/echolocal/internal/ui/theme"
	sharedlib "github.com/ygelfand/libcountertop/pkg/display/widgets"
)

type paint = sharedlib.Paint

func painted(key string, w, h int, fill theme.Color, draw paint) string {
	return sharedlib.Painted(key, w, h, fill, presentation.Palette(), draw)
}

var nrgba = sharedlib.NRGBA
var picture = sharedlib.Picture
