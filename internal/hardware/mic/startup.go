package mic

import "github.com/ygelfand/echolocal/internal/component"

// Startup reports actual acquisition without opening the audio device again.
func (s *Source) Startup() component.Progress {
	s.devMu.Lock()
	open := s.pcm != nil
	s.devMu.Unlock()
	if open {
		return component.Progress{Done: true}
	}
	if state := s.startup.Load(); state != nil && state.Failed {
		return *state
	}
	return component.Progress{Doing: "taking the microphones"}
}
