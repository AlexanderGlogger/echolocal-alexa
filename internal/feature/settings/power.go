//go:build board_checkers || board_cronos || board_rook

package settings

import (
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"

	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/update"
)

func powerPage() *shell.Page {
	return sharedsettings.PowerPage(shell.Get().Push, restartService, rebootDevice)
}

var confirmPage = sharedsettings.ConfirmPage

func restartService() { update.Restart("asked for on the panel") }

func rebootDevice() { update.Reboot("asked for on the panel") }
