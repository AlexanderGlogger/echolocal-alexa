package installer

import (
	"context"
	"fmt"
	"time"

	"github.com/ygelfand/libcountertop/pkg/host/adb"

	"github.com/ygelfand/echolocal/internal/layout"
)

const residentTimeout = 60 * time.Second

var rebootSteps = []step{
	{"reboot the device", rebootDevice},
	{"wait for Android", awaitAndroid},
	{"echod started on its own", awaitResident},
}

func RebootAndWait(ctx context.Context, d *adb.Device, report Reporter) error {
	r := &run{d: d, ctx: ctx}
	r.board()
	return execute(ctx, rebootSteps, r, report)
}

func rebootDevice(r *run) (string, bool, error) {
	return "", false, r.d.Reboot("")
}

func awaitAndroid(r *run) (string, bool, error) {
	ctx, cancel := context.WithTimeout(r.ctx, androidTimeout)
	defer cancel()

	wait := r.d.WaitBooted
	if !r.board().SignalsBoot {
		wait = r.d.WaitRooted
	}
	if err := wait(ctx); err != nil {
		return "", false, err
	}

	up, err := r.d.Uptime()
	if err != nil {
		return "booted", false, nil
	}
	return fmt.Sprintf("booted, %.0fs up", up), false, nil
}

// Non-persistent properties do not survive a reboot.
func awaitResident(r *run) (string, bool, error) {
	deadline := time.Now().Add(residentTimeout)

	for {
		state, err := r.d.Getprop(layout.StateProp)
		if err != nil {
			return "", false, err
		}
		if state == "resident" {
			started, _ := r.d.Getprop(layout.StartedProp)
			return "resident, started at uptime " + started, false, nil
		}
		if time.Now().After(deadline) {
			name := r.board().ServiceName
			svc, _ := r.d.Getprop("init.svc." + name)
			return "", false, fmt.Errorf("echod is %q %s after boot (init.svc.%s=%q)",
				state, residentTimeout, name, svc)
		}

		select {
		case <-r.ctx.Done():
			return "", false, r.ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
