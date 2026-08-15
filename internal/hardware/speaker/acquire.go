package speaker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ygelfand/libcountertop/pkg/audio/alsa"
	"github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/amazon"
	"github.com/ygelfand/echolocal/internal/android/prop"
)

const MediaService = "media"

const (
	acquireRetry    = 250 * time.Millisecond
	acquireAttempts = 8
)

// mediaserver retakes the device within a few hundred ms; leaving it stopped trips the framework watchdog.
func (p *Player) Start(context.Context) (err error) {
	defer func() {
		state := component.Progress{Done: err == nil}
		if err != nil {
			state.Failed = true
			state.Doing = err.Error()
		}
		p.startup.Store(&state)
	}()
	if amazon.Enabled() {
		return startAndroid()
	}
	err = p.open()
	if err == nil || !errors.Is(err, alsa.ErrBusy) {
		return err
	}

	slog.Warn("playback device busy, stopping "+MediaService+" to take it", "err", err)
	if err := prop.Stop(MediaService); err != nil {
		return err
	}
	defer func() {
		if err := prop.Start(MediaService); err != nil {
			slog.Error("restarting "+MediaService+" failed", "err", err)
		}
	}()

	for range acquireAttempts {
		time.Sleep(acquireRetry)

		err = p.open()
		if err == nil {
			slog.Info("playback device acquired")
			return nil
		}
		if !errors.Is(err, alsa.ErrBusy) {
			return err
		}
	}
	return err
}

func Acquire() (*Player, error) {
	p := New()
	if err := p.Start(context.Background()); err != nil {
		return nil, err
	}
	return p, nil
}
