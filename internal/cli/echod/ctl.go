package echod

import (
	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/feature/control"
)

func newCtlCmd() *cobra.Command {
	return harness.ClientCommand(harness.ClientOptions{Name: "echod", Socket: control.Socket})
}
