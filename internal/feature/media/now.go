package media

import (
	"fmt"

	sharedsource "github.com/ygelfand/libcountertop/pkg/media/source"
)

type Source = sharedsource.Source
type Details = sharedsource.Details
type Controller = sharedsource.Controller
type Now = sharedsource.Now
type Track = sharedsource.Track
type Kind = sharedsource.Kind
type Controls = sharedsource.Controls
type Seeker = sharedsource.Seeker
type Opener = sharedsource.Opener

const (
	FromQueue     = sharedsource.FromQueue
	FromBluetooth = sharedsource.FromBluetooth
	FromGroup     = sharedsource.FromGroup
	FromCast      = sharedsource.FromCast
	CanPause      = sharedsource.CanPause
	CanStop       = sharedsource.CanStop
	CanNext       = sharedsource.CanNext
	CanPrevious   = sharedsource.CanPrevious
)

// Now is what the device is playing, whoever started it.
func (p *Player) Now() Now {
	if d := p.source(); d != nil {
		return d.Now()
	}

	playing, paused := p.stream.Playing()
	return Now{Playing: playing, Paused: paused, Can: CanPause | CanStop}
}

// From is the kind holding the card, which is the queue when nothing else has it.
func (p *Player) From() Kind {
	if d := p.source(); d != nil {
		return d.Kind()
	}
	return FromQueue
}

// Named is what to call whoever holds the card, and empty where it has no name.
func (p *Player) Named() string {
	if d := p.source(); d != nil {
		return d.Label()
	}
	return ""
}

// Next and Previous reach whoever is playing, and do nothing where it cannot skip.
func (p *Player) Next() {
	if d := p.source(); d != nil {
		d.Next()
	}
}

func (p *Player) Previous() {
	if d := p.source(); d != nil {
		d.Previous()
	}
}

// Began hands the card to a source that has started playing. One of a different kind stops whoever
// held it before.
func (p *Player) Began(s Source) {
	replaced, fresh, displaced := p.sourceOwner().Begin(s)
	p.refresh()
	if fresh {
		p.Begun.Emit(s)
	}
	if displaced {
		replaced.Stop()
	}
}

// Ended says a source of this kind has stopped, so the next one of its kind counts as new.
func (p *Player) Ended(s Source) { p.sourceOwner().End(s) }

// Release gives the card back, if s still holds it.
func (p *Player) Release(s Source) {
	if p.sourceOwner().Release(s) {
		p.refresh()
	}
}

// Holder is whoever holds the card, nil when the local queue does.
func (p *Player) Holder() Source { return p.source() }

// Sourced names whoever holds the card, for a log, and whether anything does.
func (p *Player) Sourced() (string, bool) {
	s, held := p.sourceOwner().Get()
	if held {
		return fmt.Sprintf("%T", s), true
	}
	return "the local queue", false
}
func kindOf(s Source) Kind {
	if s != nil {
		return s.Kind()
	}
	return FromQueue
}
