package echoctl

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ygelfand/libcountertop/pkg/host/adb"
	"github.com/ygelfand/libcountertop/pkg/host/prompt"

	"github.com/ygelfand/echolocal/internal/host/installer"
)

func approveFlash(
	ctx context.Context,
	out io.Writer,
	d *adb.Device,
	state installer.BootState,
	image string,
	assumeYes bool,
) (bool, error) {
	if assumeYes {
		return true, nil
	}
	if !prompt.IsTerminal() {
		return false, errors.New(
			"writing the boot partition needs confirmation: run this on a terminal, or pass --yes",
		)
	}

	fmt.Fprintf(
		out,
		"\n%s\n",
		styleTitle.Render("This takes the device over, and undoing it needs a reflash"),
	)
	fmt.Fprintf(out, "  device     %s (%s)\n", d.Serial(), state.Summary)
	if image != "" {
		fmt.Fprintf(out, "  partition  %s\n", state.Partition)
		fmt.Fprintf(out, "  image      %s\n", image)
	}

	return prompt.Typed(ctx, out, "Type yes to continue", "yes")
}
