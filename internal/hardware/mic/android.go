package mic

import (
	"context"
	"encoding/binary"
	"errors"
	"log/slog"
	"time"

	"github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/amazon"
)

// On an installation with the Android media bridge, Amazon's Pryon detector owns the firmware's one
// capture stream and the bridge reads a second copy of it, already mixed to 16 kHz mono. There is no
// array to beamform or loopback to cancel against, so those stages are skipped and the frames join
// the shared path at leveling.

func startAndroid() error {
	if !amazon.Get().Connected() {
		return errors.New("mic: Android media helper is not connected")
	}
	slog.Info("capture using Android media helper", "rate", Rate, "channels", 1, "bits", 16)
	return nil
}

// androidStartup is Startup on the bridge, which leaves no device here to look at: what the last
// Start reported is all there is.
func (s *Source) androidStartup() component.Progress {
	if state := s.startup.Load(); state != nil {
		return *state
	}
	return component.Progress{Doing: "waiting for the Android media bridge"}
}

func (s *Source) runAndroid(ctx context.Context) error {
	chunks := make(chan []byte, 16)
	unlisten := amazon.Get().ListenAudio(func(chunk []byte) {
		select {
		case chunks <- chunk:
		default:
			s.dropped.Add(1)
		}
	})
	defer unlisten()

	report := time.NewTicker(dropReport)
	defer report.Stop()

	var frames reframer
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-report.C:
			s.reportDrops()
		case chunk := <-chunks:
			frames.push(chunk, s.broadcastMono)
		}
	}
}

// broadcastMono hands on a frame that arrived already mixed. Nothing here cancels its echo, so the
// leveler hears it as it is.
func (s *Source) broadcastMono(frame []int16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliver(frame, frame)
}

// reframer cuts the bridge's byte stream into FrameSamples frames. The bridge forwards whatever each
// socket read returned, and everything downstream counts time in whole frames.
type reframer struct{ pending []byte }

func (r *reframer) push(chunk []byte, emit func([]int16)) {
	const frameBytes = FrameSamples * 2

	r.pending = append(r.pending, chunk...)
	for len(r.pending) >= frameBytes {
		frame := make([]int16, FrameSamples)
		for i := range frame {
			frame[i] = int16(binary.LittleEndian.Uint16(r.pending[i*2:]))
		}
		emit(frame)
		r.pending = r.pending[frameBytes:]
	}
	// Keep the remainder at the front of a buffer that does not grow without bound.
	r.pending = append(r.pending[:0:0], r.pending...)
}
