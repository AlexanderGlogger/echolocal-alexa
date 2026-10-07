//go:build board_checkers || board_cronos || board_rook

package settings

import (
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/update"
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
)

// powerPage is the two ways to make the device go away and come back.
//
// On the panel as well as in Home Assistant, because the thing most likely to need restarting is
// what talks to Home Assistant.
func powerPage() *shell.Page {
	return sharedsettings.PowerPage(shell.Get().Push, restartService, rebootDevice)
}

// confirmPage asks once more before something that takes the device away.
//
// A second screen rather than a row that acts: this list is scrolled past with a thumb, and neither
// of these is worth doing by accident. The header is the way out, as on every other page.
var confirmPage = sharedsettings.ConfirmPage

func restartService() { update.Restart("asked for on the panel") }

func rebootDevice() { update.Reboot("asked for on the panel") }
