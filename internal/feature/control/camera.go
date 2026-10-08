//go:build board_checkers || board_cronos

package control

import (
	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/cameracmd"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/livecam"
	"github.com/ygelfand/echolocal/internal/feature/vision"
)

func init() {
	boardCommands = append(boardCommands, func(t *harness.Tree) { t.Top(cameraCommand()) })
	boardSettings = append(boardSettings, cameraSettings)
}

func cameraCommand() *cobra.Command {
	camera := harness.Group("camera", "The camera", "")
	camera.AddCommand(harness.Says(&cobra.Command{
		Use:   "still [path]",
		Short: "Take the picture Home Assistant would get, to " + cameracmd.DefaultStill + " or path",
		Args:  cobra.MaximumNArgs(1),
	}, func(args []string) (string, error) { return cameracmd.Still(vision.Get().Still, args) }))
	return camera
}

func cameraSettings() []harness.Setting[config.Config] {
	return harness.Map(
		cameracmd.Rows(livecam.Get().Camera),
		func(c config.Config) schema.Camera { return c.Camera },
	)
}
