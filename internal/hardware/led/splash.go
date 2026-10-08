package led

import (
	"context"
	"time"
)

const (
	WalkStep = 250 * time.Millisecond

	SplashConfirm = 2 * time.Second
)

func Splash(ctx context.Context, r *Ring, ready func() bool) error {
	if ready != nil {
		if err := until(ctx, r, walk(HomeAssistant), ready); err != nil {
			return err
		}
	}
	if err := play(ctx, r, SplashConfirm, comet(Palette{HomeAssistant})); err != nil {
		return err
	}

	fade, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	defer cancel()
	return fadeOut(fade, r, 250*time.Millisecond)
}

func until(ctx context.Context, r *Ring, frame Frame, ready func() bool) error {
	t := time.NewTicker(FrameInterval)
	defer t.Stop()

	start := time.Now()
	for !ready() {
		if err := r.SetSegments(frame(time.Since(start))); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
	return nil
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
