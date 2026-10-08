package echod

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ygelfand/echolocal/internal/update"
)

// The root source is /dev/root with no node: mount(8) fails, mount(2) with no source works.
func newRemountCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "remount rw|ro",
		Short:     "Make the system partition writable, or put it back",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"rw", "ro"},
		RunE: func(_ *cobra.Command, args []string) error {
			rw := args[0] == "rw"
			if err := update.Remount(rw); err != nil {
				return err
			}
			fmt.Println(args[0])
			return nil
		},
	}
}
