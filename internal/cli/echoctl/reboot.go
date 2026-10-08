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

type rebootChoice int

const (
	rebootAsk rebootChoice = iota
	rebootYes
	rebootNo
)

func rebootChoiceOf(yes, no bool) rebootChoice {
	switch {
	case no:
		return rebootNo
	case yes:
		return rebootYes
	}
	return rebootAsk
}

func offerReboot(
	ctx context.Context,
	out io.Writer,
	d *adb.Device,
	settles bool,
	choice rebootChoice,
) error {
	switch choice {
	case rebootNo:
		return nil

	case rebootAsk:
		if !settles {
			return nil
		}
		if !prompt.IsTerminal() {
			fmt.Fprintf(out, "%s\n", styleSkip.Render(
				"• some of this takes effect on the next boot; pass --reboot to have it done here"))
			return nil
		}

		yes, err := prompt.Confirm(ctx, out,
			"Reboot now? Some of what was installed only takes effect on the next boot.")
		if err != nil && !errors.Is(err, prompt.ErrCanceled) {
			return err
		}
		if !yes {
			fmt.Fprintf(
				out,
				"%s\n",
				styleSkip.Render("• not rebooting; the rest lands whenever it next boots"),
			)
			return nil
		}
	}

	return render(ctx, out, "Rebooting", "✓ device came back and echod started on its own",
		func(report installer.Reporter) error {
			return installer.RebootAndWait(ctx, d, report)
		})
}
