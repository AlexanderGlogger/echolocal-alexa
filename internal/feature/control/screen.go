//go:build board_checkers || board_cronos || board_rook

package control

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/spf13/cobra"
	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/hardware/display"
)

func showing(t *harness.Tree) {
	shot := &cobra.Command{
		Use:   "shot <path>",
		Short: "Write what the panel is showing to a PNG",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := save(args[0]); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), args[0])
			return nil
		},
	}

	open, stack, ready := panelCommands()
	t.Add(harness.DisplayTask, shot, thumbsCommand(), open, stack)
	t.Add(harness.DeviceTask, ready)
	t.Top(setCommand())
}

func save(path string) error {
	img, err := composed()
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func composed() (image.Image, error) {
	c := display.Get().Helper()
	if c == nil {
		return drawn()
	}
	pix, nw, nh, err := c.ScreenRead()
	if err != nil {
		return nil, err
	}
	rot := display.Get().Orientation()
	w, h := rot.Size(nw, nh)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			px, py := rot.Project(nw, nh, x, y)
			copy(img.Pix[img.PixOffset(x, y):][:4], pix[(py*nw+px)*4:][:4])
		}
	}
	return img, nil
}

func drawn() (image.Image, error) {
	pixels, w, h, err := display.Get().Shot()
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := range w * h {
		img.Set(i%w, i/w, color.RGBA{R: pixels[i*3], G: pixels[i*3+1], B: pixels[i*3+2], A: 0xff})
	}
	return img, nil
}
