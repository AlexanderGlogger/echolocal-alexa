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

func (p *Player) Now() Now {
	if d := p.source(); d != nil {
		return d.Now()
	}

	playing, paused := p.stream.Playing()
	return Now{Playing: playing, Paused: paused, Can: CanPause | CanStop}
}

func (p *Player) From() Kind {
	if d := p.source(); d != nil {
		return d.Kind()
	}
	return FromQueue
}

func (p *Player) Named() string {
	if d := p.source(); d != nil {
		return d.Label()
	}
	return ""
}

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

func (p *Player) Ended(s Source) { p.sourceOwner().End(s) }

func (p *Player) Release(s Source) {
	if p.sourceOwner().Release(s) {
		p.refresh()
	}
}

func (p *Player) Holder() Source { return p.source() }

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
