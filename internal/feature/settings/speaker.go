//go:build with_screen

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/a2dp"
)

func init() {
	features = append(features, feature{
		key: "features.speaker",
		on:  func(c config.Config) bool { return c.Bluetooth.Speaker },
		set: func(on bool) { a2dp.Get().SetEnabled(on) },
	})
}
