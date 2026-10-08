//go:build board_checkers || board_cronos

package settings

import (
	"github.com/ygelfand/echolocal/internal/feature/shell"
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
