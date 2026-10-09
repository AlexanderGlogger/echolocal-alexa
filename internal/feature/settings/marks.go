//go:build with_screen

package settings

import (
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
	tesseraview "github.com/ygelfand/libcountertop/pkg/tessera/widgets"

	echoui "github.com/ygelfand/echolocal/internal/ui"
)

func modeMarks() sharedsettings.ModeMarks {
	return sharedsettings.ModeMarks{
		Tessera: tesseraview.Mark,
		Native:  sharedsettings.Themed(echoui.Logo().Light, echoui.Logo().Dark),
	}
}
