//go:build board_checkers || board_cronos || board_rook

package gui

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/libcountertop/pkg/display/widgets"
)

var presentation = widgets.NewUI(config.ScreenSection)
