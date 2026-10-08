package mic

type Echo struct {
	Mic, Ref, Out []int16
	Sounding      bool
}

func (s *Source) ListenEcho() (<-chan Echo, func()) {
	ch := make(chan Echo, 64)

	s.mu.Lock()
	id := s.next
	s.next++
	s.echoes[id] = ch
	s.mu.Unlock()

	return ch, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if c, ok := s.echoes[id]; ok {
			delete(s.echoes, id)
			close(c)
		}
	}
}
