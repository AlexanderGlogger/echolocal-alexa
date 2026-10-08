package mic

import (
	"sync"
	"time"
)

const History = 250 * time.Millisecond

const historySamples = Rate

type history struct {
	mu     sync.Mutex
	buf    [historySamples]int16
	at     int
	filled bool
}

func (h *history) add(frame []int16) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, s := range frame {
		h.buf[h.at] = s
		h.at++
		if h.at == len(h.buf) {
			h.at, h.filled = 0, true
		}
	}
}

func (h *history) recent(d time.Duration) []int16 {
	want := int(d/time.Millisecond) * Rate / 1000

	h.mu.Lock()
	defer h.mu.Unlock()

	have := len(h.buf)
	if !h.filled {
		have = h.at
	}
	want = min(want, have)

	out := make([]int16, 0, want)
	start := h.at - want
	if start < 0 {
		start += len(h.buf)
	}
	for i := range want {
		out = append(out, h.buf[(start+i)%len(h.buf)])
	}
	return out
}

func (s *Source) remember(frame []int16) { s.history.add(frame) }

func (s *Source) Recent(d time.Duration) []int16 { return s.history.recent(d) }
