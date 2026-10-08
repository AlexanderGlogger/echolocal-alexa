//go:build board_checkers || board_cronos || board_rook

package control

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygelfand/libcountertop/pkg/display/panel"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/screencmd"

	pages "github.com/ygelfand/echolocal/internal/feature/settings"
	screens "github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/splash"
	"github.com/ygelfand/echolocal/internal/hardware/display"
)

const (
	readyWait  = 90 * time.Second
	readyEvery = 200 * time.Millisecond
)

func panelCommands() (open, stack, ready *cobra.Command) {
	open = &cobra.Command{
		Use:   "open <page>",
		Short: "Put a settings page up: " + strings.Join(pages.Pages(), ", "),
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if !pages.OpenPage(args[0]) {
				return fmt.Errorf(
					"no page %q, want one of %s",
					args[0],
					strings.Join(pages.Pages(), ", "),
				)
			}
			return nil
		},
	}

	stack = &cobra.Command{
		Use:   "stack",
		Short: "The screens up, top first, and which are held",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), strings.TrimRight(screencmd.Stack(screens.Get()), "\n"))
			return nil
		},
	}

	ready = &cobra.Command{
		Use:   "ready [milliseconds]",
		Short: "Wait until the boot screen lets go of the panel, and say what has it",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wait := readyWait
			if len(args) > 0 {
				ms, err := strconv.Atoi(args[0])
				if err != nil || ms < 0 {
					return fmt.Errorf("milliseconds must be a number")
				}
				wait = time.Duration(ms) * time.Millisecond
			}
			for deadline := time.Now().Add(wait); ; time.Sleep(readyEvery) {
				if splash.Get().Finished() {
					fmt.Fprintln(cmd.OutOrStdout(), "screen")
					return nil
				}
				if at, showing := display.Get().Showing(); showing && at < panel.PriorityBoot {
					fmt.Fprintln(cmd.OutOrStdout(), shown(at))
					return nil
				}
				if time.Now().After(deadline) {
					return fmt.Errorf("still starting after %s", wait)
				}
			}
		},
	}

	return open, stack, ready
}

func shown(at panel.Priority) string {
	switch at {
	case panel.PriorityDashboard:
		return "dashboard"
	case panel.PrioritySetup:
		return "setup"
	case panel.PriorityUI:
		return "screen"
	case panel.PriorityNotice:
		return "notice"
	case panel.PriorityAlert:
		return "alert"
	}
	return "boot"
}
