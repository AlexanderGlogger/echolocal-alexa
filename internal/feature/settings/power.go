//go:build !board_biscuit

package settings

import (
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
	sharedshell "github.com/ygelfand/libcountertop/pkg/display/shell"

	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/update"
)

func powerPage() *sharedshell.Page {
	return sharedsettings.PowerPage(shell.Get().Push, restartService, rebootDevice)
}

func restartService() { update.Restart("asked for on the panel") }

func rebootDevice() { update.Reboot("asked for on the panel") }
