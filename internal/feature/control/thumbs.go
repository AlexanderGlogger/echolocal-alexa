//go:build board_checkers || board_cronos || board_rook

package control

import (
	"errors"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ygelfand/libcountertop/pkg/display/visual"
	"github.com/ygelfand/libcountertop/pkg/runtime/control/screencmd"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/display"
	"github.com/ygelfand/echolocal/internal/hardware/gpu"
)

const (
	DefaultThumbs = "/data/local/tmp/echolocal-thumbs"
	thumbWidth    = 384
)

func thumbsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "thumbs [dir]",
		Short: "Render the visual picker's thumbnails on this panel",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := DefaultThumbs
			if len(args) > 0 {
				dir = args[0]
			}
			n, err := thumbs(dir)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "wrote %d thumbnails to %s\n", n, dir)
			return nil
		},
	}
}

func thumbs(dir string) (int, error) {
	w, h := display.Get().Size()
	if w == 0 {
		return 0, errors.New("there is no panel to size the thumbnails by")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, err
	}

	n := 0
	for _, k := range visual.Built() {
		img, err := screencmd.Render(
			gpu.Get(),
			k,
			w,
			h,
			screencmd.ThumbFrames,
			config.Defaults().Visual.Label,
		)
		if err != nil {
			return n, fmt.Errorf("%s: %w", k, err)
		}
		f, err := os.Create(filepath.Join(dir, string(k)+".jpg"))
		if err != nil {
			return n, err
		}
		err = jpeg.Encode(
			f,
			screencmd.Shrink(img, thumbWidth, thumbWidth*h/w),
			&jpeg.Options{Quality: 82},
		)
		f.Close()
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
