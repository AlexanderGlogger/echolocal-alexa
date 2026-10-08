package echoctl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ygelfand/echolocal/internal/host/device"
)

var ErrCancelled = errors.New("cancelled")

func connect(ctx context.Context, out io.Writer, serial string) (*device.Device, error) {
	target, err := resolveSerial(ctx, out, serial, device.List, "")
	if err != nil {
		return nil, err
	}
	return device.Connect(target)
}

func attach(ctx context.Context, out io.Writer, serial string) (*device.Device, error) {
	for {
		d, err := attachOnce(ctx, out, serial)
		if !errors.Is(err, device.ErrUnreachable) || !isTerminal() {
			return d, err
		}
		if err := offerRecovery(ctx, out); err != nil {
			return nil, err
		}
	}
}

const noneOfThese = "None of these"

func attachOnce(ctx context.Context, out io.Writer, serial string) (*device.Device, error) {
	target, err := resolveSerial(ctx, out, serial, device.ListAny, noneOfThese)
	if err != nil {
		return nil, err
	}
	return device.AttachAny(target)
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
			return ErrCancelled
		case <-tick.C:
		}
	}
}

func listAny() []device.Info {
	devices, err := device.ListAny()
	if err != nil {
		return nil
	}
	return devices
}

func resolveSerial(
	ctx context.Context,
	out io.Writer,
	serial string,
	list func() ([]device.Info, error),
	other string,
) (string, error) {
	if serial != "" {
		return serial, nil
	}
	if !isTerminal() {
		return "", nil
	}

	devices, err := list()
	if err != nil {
		return "", err
	}
	if len(devices) == 0 {
		return "", device.ErrUnreachable
	}
	return pickDevice(ctx, out, devices, other)
}

func pickDevice(
	ctx context.Context,
	out io.Writer,
	devices []device.Info,
	other string,
) (string, error) {
	chosen, err := choose(ctx, out, "Select a device", devices,
		func(d device.Info) string { return fmt.Sprintf("%s  %s", d, d.Serial) }, other)
	if err != nil {
		return "", err
	}
	if chosen.Serial == "" {
		return "", device.ErrUnreachable
	}
	return chosen.Serial, nil
}
