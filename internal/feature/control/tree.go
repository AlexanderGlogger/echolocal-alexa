package control

import (
	"fmt"

	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
)

func (c *Control) tree() *cobra.Command {
	root := harness.NewRoot()

	root.AddCommand(telling()...)
	root.AddCommand(showing()...)
	root.AddCommand(hearing()...)
	root.AddCommand(cancelling()...)
	root.AddCommand(waiting())
	root.AddCommand(c.inputEngine().Commands()...)
	for _, more := range boardCommands {
		root.AddCommand(more())
	}
	harness.Organize(root)
	return root
}

var boardCommands []func() *cobra.Command

func telling() []*cobra.Command {
	commands := harness.DeviceCommands(deviceInfo)

	logs := &cobra.Command{
		Use:       "log [debug|info|warn|error]",
		Short:     "Read or change how much goes to logcat, until echod restarts",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"debug", "info", "warn", "error"},
		RunE: func(cmd *cobra.Command, args []string) error {
			level, err := logLevel(args)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), level)
			return nil
		},
	}

	return append(commands, logs)
}

func waiting() *cobra.Command { return harness.WaitCommand() }
