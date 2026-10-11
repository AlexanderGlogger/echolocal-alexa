package speaker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/amazon"
)

// On an installation with the Android media bridge, Amazon's Pryon detector shares the firmware audio
// service, so playback goes through AudioTrack in the bridge rather than to the codec directly. The
// queue, mixing and volume are unchanged: only where each filled period is written differs.

// androidTuned reports whether Fire OS tunes the speaker itself. Its audio service runs the same
// vendor EQ and MBCL from asp.VendorDir on everything AudioTrack plays, so running the chain here
// too would apply the tuning twice: muffled, with the compressors lifting the noise floor.
func androidTuned() bool { return amazon.Enabled() }

func startAndroid() error {
	if !amazon.Get().Connected() {
		return errors.New("speaker: Android media helper is not connected")
	}
	slog.Info(
		"playback using Android media helper",
		"rate",
		Rate,
		"channels",
		Channels,
		"bits",
		Bits,
	)
	return nil
}

// androidStartup is Startup on the bridge, which leaves no device here to look at: what the last
// Start reported is all there is.
func (p *Player) androidStartup() component.Progress {
	if state := p.startup.Load(); state != nil {
		return *state
	}
	return component.Progress{Doing: "waiting for the Android media bridge"}
}

func (p *Player) runAndroid(ctx context.Context) error {
	buf := make([]byte, period*Channels*Bits/8)
	slog.Info("playback path up", "output", "android-media")
	p.OnOutput.Emit(p.Output())

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		p.fill(buf)
		if err := amazon.Get().Play(buf); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("speaker: Android media write: %w", err)
		}
		p.written.Add(period)
	}
}

func closeAndroid() error {
	if amazon.Get().Connected() {
		return amazon.Get().StopPlayback()
	}
	return nil
}

// available reports whether anything will play what is queued.
func (p *Player) available() bool {
	if amazon.Enabled() {
		return amazon.Get().Connected()
	}
	pb, _ := p.device()
	return pb != nil
}
