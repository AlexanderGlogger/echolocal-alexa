package boot

import (
	"context"

	"github.com/ygelfand/libcountertop/pkg/runtime/safe"

	"github.com/ygelfand/echolocal/internal/hardware/led"
)

func startSplash(ctx context.Context, leds *led.Driver, ready func() bool) {
	if !leds.Present() {
		return
	}

	safe.Go("splash", func() {
		claim := leds.Claim(led.PriorityBoot)
		defer claim.Release()

		done := make(chan struct{}, 1)
		claim.Show(led.Content{Animate: func(actx context.Context, r *led.Ring) error {
			err := led.Splash(actx, r, ready)
			select {
			case done <- struct{}{}:
			default:
			}
			return err
		}})

		select {
		case <-done:
		case <-ctx.Done():
		}
	})
}
