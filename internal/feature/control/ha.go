//go:build board_checkers || board_cronos || board_rook

package control

import (
	"github.com/spf13/cobra"
	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	sharedcmd "github.com/ygelfand/libcountertop/pkg/runtime/control/homecmd"
)

func homeAssistant() []*cobra.Command { return sharedcmd.Commands(homeassistant.Get()) }
