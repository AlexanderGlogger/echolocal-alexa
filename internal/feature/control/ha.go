//go:build board_checkers || board_cronos || board_rook

package control

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/homeassistant"
	"github.com/ygelfand/echolocal/internal/feature/homecontrol"
)

const haWait = 10 * time.Second

func init() {
	boardCommands = append(boardCommands, haCommand)
	boardSettings = append(boardSettings, homeRows)
}

func haCommand() *cobra.Command {
	ha := &cobra.Command{
		Use:   "ha",
		Short: "Ask Home Assistant",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
	}
	ha.AddCommand(
		&cobra.Command{
			Use:   "access",
			Short: "Whether Home Assistant answers actions from this device",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), homeassistant.Get().Access().String())
				return nil
			},
		},
		&cobra.Command{
			Use:   "entities [DOMAIN...]",
			Short: "The entities Home Assistant has, in every domain or the ones named",
			RunE: func(cmd *cobra.Command, args []string) error {
				ctx, cancel := context.WithTimeout(context.Background(), haWait)
				defer cancel()
				got, err := homeassistant.Get().Entities(ctx, homeassistant.Filter{Domains: args})
				if err != nil {
					return err
				}
				var out strings.Builder
				for _, e := range got {
					fmt.Fprintf(&out, "%s\t%s\t%s\t%s\n", e.ID, e.State, e.Area, e.Name)
				}
				fmt.Fprint(cmd.OutOrStdout(), out.String())
				return nil
			},
		},
		&cobra.Command{
			Use:   "labels",
			Short: "The labels Home Assistant has",
			Args:  cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				ctx, cancel := context.WithTimeout(context.Background(), haWait)
				defer cancel()
				got, err := homeassistant.Get().Labels(ctx)
				if err != nil {
					return err
				}
				var out strings.Builder
				for _, l := range got {
					fmt.Fprintf(&out, "%s\t%s\n", l.ID, l.Name)
				}
				fmt.Fprint(cmd.OutOrStdout(), out.String())
				return nil
			},
		},
		&cobra.Command{
			Use:   "light on|off ENTITY",
			Short: "Turn a light on or off",
			Args:  cobra.ExactArgs(2),
			RunE: func(cmd *cobra.Command, args []string) error {
				action := map[string]string{"on": "light.turn_on", "off": "light.turn_off"}[args[0]]
				if action == "" {
					return fmt.Errorf("on or off, not %q", args[0])
				}
				ctx, cancel := context.WithTimeout(context.Background(), haWait)
				defer cancel()
				if err := homeassistant.Get().
					Call(ctx, action, map[string]string{"entity_id": args[1]}); err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), args[1]+" "+args[0])
				return nil
			},
		},
	)
	return ha
}

func homeRows() []setting {
	var rows []setting
	for _, s := range homecontrol.Selections {
		rows = append(rows, setting{
			name: "home.control." + s.Key,
			says: func(c config.Config) string { return onOff(c.Home.Control[s.Key]) },
			use:  toggle(s.SetControlled),
		}, setting{
			name: "home.group." + s.Key,
			says: func(c config.Config) string { return strconv.Itoa(c.Home.Grouped(s.Key)) },
			use:  number(0, config.HomeGroupMost, s.SetGroupFrom),
		})
	}
	rows = append(rows, setting{
		name: "home.enabled",
		says: func(c config.Config) string { return onOff(c.Home.Enabled) },
		use:  toggle(homecontrol.SetEnabled),
	})
	return append(rows, setting{
		name: "home.combine",
		says: func(c config.Config) string { return onOff(c.Home.Combine) },
		use:  toggle(homecontrol.SetCombined),
	})
}
