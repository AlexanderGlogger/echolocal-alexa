package echoctl

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ygelfand/echolocal/internal/host/device"
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
	d *device.Device,
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
		if !isTerminal() {
			fmt.Fprintf(out, "%s\n", styleSkip.Render(
				"• some of this takes effect on the next boot; pass --reboot to have it done here"))
			return nil
		}

		yes, err := confirm(ctx, out,
			"Reboot now? Some of what was installed only takes effect on the next boot.")
		if err != nil && !errors.Is(err, ErrCancelled) {
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
