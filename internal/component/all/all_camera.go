//go:build !board_biscuit

package all

import (
	_ "github.com/ygelfand/echolocal/internal/feature/livecam"
	_ "github.com/ygelfand/echolocal/internal/feature/rtspd"
	_ "github.com/ygelfand/echolocal/internal/feature/vision"
)
