package voice

import (
	"time"

	"github.com/ygelfand/echolocal/internal/feature/light"
	"github.com/ygelfand/echolocal/internal/hardware/led"
)

type turnEffects struct{ claim *led.Claim }

func (e *turnEffects) Begin() { e.claim = led.Get().Claim(led.PriorityTurn) }
func (e *turnEffects) End()   { e.claim.Release() }
func (e *turnEffects) Clear() { e.claim.Clear() }

func (e *turnEffects) Elsewhere() { led.Get().Busy().Flash(led.WorkElsewhere, 900*time.Millisecond) }
func (e *turnEffects) Play(name string, reversed bool) {
	base := light.Get().Base()
	if reversed {
		e.claim.PlayReversed(name, base)
	} else {
		e.claim.Play(name, base)
	}
}
