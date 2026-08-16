package led

import (
	"context"
	"time"
)

const (
	WalkStep = 250 * time.Millisecond

	SplashConfirm = 2 * time.Second

	// SplashWait bounds the wait for Home Assistant. A ring still walking long after boot looks like
	// recovery or a boot loop, so a device nobody has subscribed to fades out instead.
	SplashWait = 60 * time.Second
)

func Splash(ctx context.Context, r *Ring, ready func() bool) error {
	return splash(ctx, r, ready, SplashWait, SplashConfirm, 250*time.Millisecond)
}

func splash(
	ctx context.Context,
	r *Ring,
	ready func() bool,
	wait, confirm, fade time.Duration,
) error {
	confirmed := ready == nil
	if ready != nil {
		waiting, cancel := context.WithTimeout(ctx, wait)
		var err error
		confirmed, err = until(waiting, r, walk(HomeAssistant), ready)
		cancel()
		if err != nil {
			return err
		}
	}
	if confirmed {
		if err := play(ctx, r, confirm, comet(Palette{HomeAssistant})); err != nil {
			return err
		}
	}

	fading, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	return fadeOut(fading, r, fade)
}

// until reports whether ready came true, so a timeout does not get the confirmation comet.
func until(ctx context.Context, r *Ring, frame Frame, ready func() bool) (bool, error) {
	t := time.NewTicker(FrameInterval)
	defer t.Stop()

	start := time.Now()
	for !ready() {
		if err := r.SetSegments(frame(time.Since(start))); err != nil {
			return false, err
		}
		select {
		case <-ctx.Done():
			return false, nil
		case <-t.C:
		}
	}
	return true, nil
}

func walk(base Color) Frame {
	return func(elapsed time.Duration) []Color {
		out := make([]Color, Segments)
		out[2*int(elapsed/WalkStep)%Segments] = base
		return out
	}
}

func fadeOut(ctx context.Context, r *Ring, d time.Duration) error {
	steps := int(d / FrameInterval)
	for i := steps; i > 0; i-- {
		if err := r.SetSegments(
			around(Palette{HomeAssistant}, float64(i)/float64(steps)),
		); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return r.Off()
		case <-time.After(FrameInterval):
		}
	}
	return r.Off()
}
