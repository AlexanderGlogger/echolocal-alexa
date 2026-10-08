//go:build board_checkers || board_cronos

package control

import (
	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/callcmd"

	"github.com/ygelfand/echolocal/internal/feature/discovery"
)

func init() {
	boardCommands = append(boardCommands, func(t *harness.Tree) {
		t.Add(harness.DeviceTask, harness.Says(&cobra.Command{
			Use:   "peers",
			Short: "Other devices on the network",
			Args:  cobra.NoArgs,
		}, func([]string) (string, error) { return callcmd.Peers(discovery.Get().Peers()), nil }))
	})
}
