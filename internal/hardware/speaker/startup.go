package speaker

import (
	"github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/amazon"
)

func (s *Player) Startup() component.Progress {
	if amazon.Enabled() {
		return s.androidStartup()
	}
	s.devMu.Lock()
	open := s.pb != nil
	s.devMu.Unlock()
	if open {
		return component.Progress{Done: true}
	}
	if state := s.startup.Load(); state != nil && state.Failed {
		return *state
	}
	return component.Progress{Doing: "taking the speaker"}
}
