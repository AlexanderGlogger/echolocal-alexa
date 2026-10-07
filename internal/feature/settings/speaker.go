//go:build board_checkers || board_cronos || board_rook

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/a2dp"
	"github.com/ygelfand/libcountertop/pkg/say"
)

func init() {
	features = append(features, feature{
		label: func() string { return say.T("features.speaker") },
		hint:  func() string { return say.T("features.speaker.hint") },
		on:    func(c config.Config) bool { return c.Bluetooth.Speaker },
		set:   func(on bool) { a2dp.Get().SetEnabled(on) },
	})
}
