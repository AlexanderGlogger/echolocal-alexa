package detect

import (
	"sync"

	"github.com/ygelfand/echolocal/internal/hardware/led"
)

type wakeBusy struct {
	busy *led.Busy

	ready func(slot int) bool

	mu   sync.Mutex
	task *led.Task

	waiting bool
	pending map[int]bool
}

func newWakeBusy(busy *led.Busy, ready func(slot int) bool) *wakeBusy {
	return &wakeBusy{busy: busy, ready: ready, pending: map[int]bool{}}
}

func (w *wakeBusy) begin() {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.waiting = false
	if w.task == nil {
		w.task = w.busy.Start(led.WorkWakeWord)
	}
}

func (w *wakeBusy) waitFor(slots []int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.pending, w.waiting = map[int]bool{}, true
	for _, n := range slots {
		if !w.ready(n) {
			w.pending[n] = true
		}
	}
	w.settle()
}

func (w *wakeBusy) scored(slot int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.pending, slot)
	w.settle()
}

func (w *wakeBusy) settle() {
	if w.task == nil || !w.waiting || len(w.pending) > 0 {
		return
	}
	w.task.Done()
	w.task, w.waiting = nil, false
}
