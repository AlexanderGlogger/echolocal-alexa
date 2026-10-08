package touch

import (
	"time"

	"github.com/ygelfand/echolocal/internal/lib/input"
)

// Protocol B axes. A contact is a slot, and a slot holds a tracking id until it goes to -1.
const (
	absMTSlot       = 0x2f
	absMTPositionX  = 0x35
	absMTPositionY  = 0x36
	absMTTrackingID = 0x39

	synReport = 0x00
)

const released = -1

type Phase int

const (
	Down Phase = iota
	Move
	Up
)

func (p Phase) String() string {
	switch p {
	case Down:
		return "down"
	case Move:
		return "move"
	case Up:
		return "up"
	}
	return "unknown"
}

type Contact struct {
	Slot  int
	ID    int
	X, Y  int
	Phase Phase
	At    time.Time
}

type slot struct {
	id      int
	x, y    int
	touched bool
	changed bool

	lifting bool
}

type decoder struct {
	current int
	slots   map[int]*slot

	place func(x, y int) (int, int)
}

func (d *decoder) at(n int) *slot {
	if d.slots == nil {
		d.slots = map[int]*slot{}
	}
	if d.slots[n] == nil {
		d.slots[n] = &slot{id: released}
	}
	return d.slots[n]
}

func (d *decoder) event(e input.Event, now time.Time) []Contact {
	switch e.Type {
	case input.EvAbs:
		d.abs(e.Code, int(e.Value))
	case input.EvSyn:
		if e.Code == synReport {
			return d.report(now)
		}
	}
	return nil
}

func (d *decoder) abs(code uint16, value int) {
	switch code {
	case absMTSlot:
		d.current = value
	case absMTTrackingID:
		s := d.at(d.current)
		s.changed = true
		if value == released {
			s.lifting = true
		} else {
			s.id, s.lifting = value, false
			s.touched = false
		}
	case absMTPositionX:
		s := d.at(d.current)
		s.x, s.changed = value, true
	case absMTPositionY:
		s := d.at(d.current)
		s.y, s.changed = value, true
	}
}

func (d *decoder) report(now time.Time) []Contact {
	var out []Contact

	for n, s := range d.slots {
		if !s.changed {
			continue
		}
		s.changed = false

		x, y := s.x, s.y
		if d.place != nil {
			x, y = d.place(x, y)
		}

		switch {
		case s.lifting:
			if s.touched {
				out = append(out, Contact{Slot: n, ID: s.id, X: x, Y: y, Phase: Up, At: now})
			}
			s.id, s.touched, s.lifting = released, false, false
		case s.id == released:
		case !s.touched:
			s.touched = true
			out = append(out, Contact{Slot: n, ID: s.id, X: x, Y: y, Phase: Down, At: now})
		default:
			out = append(out, Contact{Slot: n, ID: s.id, X: x, Y: y, Phase: Move, At: now})
		}
	}
	return out
}
