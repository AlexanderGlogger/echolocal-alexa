// Package livecam supplies board-specific configuration to the shared camera-session hub.
package livecam

import (
	"errors"
	"github.com/ygelfand/echolocal/internal/android/prop"
	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/mtkcamera"
	"github.com/ygelfand/echolocal/internal/hardware/privacy"
	"github.com/ygelfand/echolocal/internal/layout"
	camerasession "github.com/ygelfand/libcountertop/pkg/camera/session"
	"log/slog"
	"time"
)

const FPS = 30
const Wait = 2 * time.Second
const Settle = 2 * time.Second

var ErrNoStill = camerasession.ErrNoStill
var ErrMuted = camerasession.ErrMuted
var ErrRestarted = errors.New("livecam: the streams changed")

type Picture = camerasession.Picture
type Session = camerasession.Session
type Size = camerasession.Size

// Sizes is each stream as it is encoded.
func Sizes() []Size {
	mw, mh := parseSize(Saved().MainSize)
	return []Size{{mw, mh}}
}

const noTurn = -1

const Keyframe = 2

func Bitrate(w, h int) int { return bitrateFor(Saved(), w, h) }

var hub = camerasession.New(camerasession.Options{
	Snapshot:     cameraSnapshot,
	Open:         func(c mtkcamera.Config) (camerasession.Transport, error) { return mtkcamera.Open(c) },
	Muted:        muted,
	Recover:      revive,
	RestartError: ErrRestarted,
	Wait:         Wait,
	Settle:       Settle,
})

func cameraSnapshot() camerasession.Snapshot {
	k := Saved()
	w, h := parseSize(k.MainSize)
	cfg := mtkcamera.Config{
		Width:    w,
		Height:   h,
		FPS:      FPS,
		Bitrate:  bitrateFor(k, w, h),
		Keyframe: k.Keyframe,
		Params:   Params(k),
		Turn:     noTurn,
	}
	return camerasession.Snapshot{
		Config:     cfg,
		Key:        shapeOf(k),
		StillExtra: time.Duration(k.Keyframe) * time.Second,
	}
}
func Join(at int) (*Session, <-chan mtkcamera.Frame, error) { return hub.Join(at) }
func Leave(s *Session, frames <-chan mtkcamera.Frame)       { hub.Leave(s, frames) }

// Sessions returns the camera owner used by streaming consumers.
func Sessions() *camerasession.Hub { return hub }

func RequestKey() error                           { return hub.RequestKey() }
func Still(within time.Duration) (Picture, error) { return hub.Still(within) }
func Muted() bool                                 { return muted() }
func muted() bool {
	if !component.Board().Has(board.MicMutesCamera) {
		return false
	}
	m, err := privacy.Microphone()
	if err != nil {
		return false
	}
	cut, err := m.Get()
	return err == nil && cut
}

const reviveEvery = 30 * time.Second

var revived time.Time

func revive(cause error) {
	if time.Since(revived) < reviveEvery {
		return
	}
	revived = time.Now()
	slog.Warn("restarting the camera helper", "err", cause)
	if err := prop.Restart(layout.CameraService); err != nil {
		slog.Error("restarting the camera helper failed", "err", err)
	}
}
