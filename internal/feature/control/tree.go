package control

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/metrics"
	"github.com/ygelfand/echolocal/internal/layout"
)

// What the harness can do, in one place.
//
// A command carries its own description, so there is one of them and it cannot go stale against a
// help string written out beside it.
//
// Built per invocation rather than once: commands hold parsed state, and these arrive down a socket
// from whoever is poking at the device, not from a process that starts and exits.
func (c *Control) tree() *cobra.Command {
	root := &cobra.Command{
		Use:   "ctl",
		Short: "Drive the running device",
		Long:  "Everything here acts on the device as it is, now.",

		SilenceUsage:  true,
		SilenceErrors: true,

		// A bare invocation is not an error, and neither is an empty line on the socket.
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}

	// Nothing down this socket has a shell to complete for.
	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(telling()...)
	root.AddCommand(showing()...)
	root.AddCommand(hearing()...)
	root.AddCommand(cancelling()...)
	root.AddCommand(waiting(), pressing())
	for _, more := range boardCommands {
		root.AddCommand(more())
	}
	return root
}

var boardCommands []func() *cobra.Command

// telling is what the device will say about itself.
func telling() []*cobra.Command {
	version := &cobra.Command{
		Use:   "version",
		Short: "What this build is",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			b := component.Board()
			fmt.Fprintf(cmd.OutOrStdout(), "version %s\ncommit  %s\nbuilt   %s\nboard   %s\n",
				layout.Version, layout.GitCommit, layout.BuildDate, b.Codename)
			return nil
		},
	}

	state := &cobra.Command{
		Use:   "state",
		Short: "What the device is doing",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			d := config.Get().Device
			up := time.Duration(metrics.Uptime() * float64(time.Second)).Round(time.Second)

			fmt.Fprintf(cmd.OutOrStdout(), "name    %s\naddr    %s\nmodel   %s\nuptime  %s\n",
				d.Name, d.Addr, d.Board.Model, up)
			return nil
		},
	}

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

	return []*cobra.Command{version, state, logs}
}

func waiting() *cobra.Command {
	return &cobra.Command{
		Use:   "wait <milliseconds>",
		Short: "Do nothing for a while, between the steps of a sequence",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			ms, err := strconv.Atoi(args[0])
			if err != nil || ms < 0 || ms > 60_000 {
				return fmt.Errorf("%q is not a number of milliseconds up to a minute", args[0])
			}
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return nil
		},
	}
}

func says(cmd *cobra.Command, do func([]string) (string, error)) *cobra.Command {
	cmd.RunE = func(c *cobra.Command, args []string) error {
		out, err := do(args)
		if out != "" {
			fmt.Fprint(c.OutOrStdout(), strings.TrimRight(out, "\n")+"\n")
		}
		return err
	}
	return cmd
}

func group(use, short, long string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("%s: %q is not one of its commands", cmd.Name(), args[0])
			}
			return cmd.Help()
		},
	}
}

func does(cmd *cobra.Command, do func([]string) error) *cobra.Command {
	cmd.RunE = func(_ *cobra.Command, args []string) error { return do(args) }
	return cmd
}
