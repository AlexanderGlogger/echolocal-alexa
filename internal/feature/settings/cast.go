//go:build board_checkers || board_cronos

package settings

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	sharedcast "github.com/ygelfand/libcountertop/pkg/display/settings/castpages"
)

func init() {
	features = append(features,
		feature{
			key:  "features.cast",
			page: castPage,
		},
	)
}

func castPage() *shell.Page { return servicePages().Cast() }

func castAppsPage() *shell.Page { return servicePages().CastApps() }

func primePage() *shell.Page { return servicePages().Prime() }

func youtubePage() *shell.Page { return servicePages().YouTube() }

func skipPage() *shell.Page { return servicePages().Sponsor() }

var toggled = sharedcast.Toggled

var skipSays = sharedcast.SkipSays

const delaySpan = config.LiveDelayMost - config.LiveDelayLeast

var delayOf = sharedcast.DelayOf

var delayLevel = sharedcast.DelayLevel

var delaySnap = sharedcast.DelaySnap

var delaySays = sharedcast.DelaySays
