//go:build !board_biscuit

package settings

import (
	"github.com/ygelfand/libcountertop/pkg/display/shell"
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
