package echolocal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ygelfand/libcountertop/pkg/host/adb"
	"github.com/ygelfand/libcountertop/pkg/host/prompt"
)

var errNotRoot = errors.New(
	"device: adbd is not running as root; write the boot image first (echolocal install --flash-only)",
)

func connect(ctx context.Context, out io.Writer, serial string) (*adb.Device, error) {
	target, err := prompt.Serial(ctx, out, serial, adb.List, "")
	if err != nil {
		return nil, err
	}
	d, err := adb.Attach(target)
	if err != nil {
		return nil, err
	}
	root, err := d.IsRoot()
	if err != nil {
		return nil, err
	}
	if !root {
		return nil, errNotRoot
	}
	return d, nil
}

func attach(ctx context.Context, out io.Writer, serial string) (*adb.Device, error) {
	for {
		d, err := attachOnce(ctx, out, serial)
		if !errors.Is(err, adb.ErrUnreachable) || !prompt.IsTerminal() {
			return d, err
		}
		if err := offerRecovery(ctx, out); err != nil {
			return nil, err
		}
	}
}

const noneOfThese = "None of these"

func attachOnce(ctx context.Context, out io.Writer, serial string) (*adb.Device, error) {
	target, err := prompt.Serial(ctx, out, serial, adb.ListAny, noneOfThese)
	if err != nil {
		return nil, err
	}
	return adb.AttachAny(target)
}
func offerRecovery(ctx context.Context, out io.Writer) error {
	fmt.Fprintf(out, "\n%s\n", styleTitle.Render("Cannot reach the device"))
	fmt.Fprintf(out, "%s\n", styleDetail.Render(
		"  If the device is up and connected but the adb patches are not installed yet,\n"+
			"  hold volume up while it powers on. On a device with a light, it is in\n"+
			"  recovery once that light is solid white."))

	already := make(map[string]bool)
	for _, d := range listAny() {
		already[d.Serial] = true
	}

	fmt.Fprintf(out, "%s\n", styleDetail.Render("  Waiting for it to appear; ctrl+c to give up."))

	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		for _, d := range listAny() {
			if !already[d.Serial] {
				fmt.Fprintf(out, "%s\n", styleDone.Render("✓ "+d.Serial+" appeared"))
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return prompt.ErrCanceled
		case <-tick.C:
		}
	}
}

func listAny() []adb.Info {
	devices, err := adb.ListAny()
	if err != nil {
		return nil
	}
	return devices
}
