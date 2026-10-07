package echod

import (
	"github.com/spf13/cobra"
	"github.com/ygelfand/echolocal/internal/feature/control"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
)

var sequence = harness.Sequence

func newCtlCmd() *cobra.Command {
	return harness.ClientCommand(harness.ClientOptions{Name: "echod", Socket: control.Socket})
}
