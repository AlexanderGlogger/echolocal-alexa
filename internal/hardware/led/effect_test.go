package led

import (
	"math"
	"slices"
	"testing"
	"time"
)

func frames(f Frame) [][]Color {
	out := make([][]Color, 0, 250)
	for step := range cap(out) {
		out = append(out, f(time.Duration(step)*FrameInterval))
	}
	return out
}

func built(t *testing.T, e Effect, base Color) Frame {
	t.Helper()

	var heard int
	frame, err := effect(e.Name, base, Room{
		Level: func() float64 {
			heard++
			return math.Abs(math.Sin(float64(heard) / 40))
		},
		Facing: func() (float64, bool) { return math.Mod(float64(heard)/120, 1), true },
	})
	if err != nil {
		t.Fatalf("building it failed: %v", err)
	}
	return frame
}

func TestEveryEffectAnimatesAndLightsTheRing(t *testing.T) {
	for _, e := range effects {
		t.Run(e.Name, func(t *testing.T) {
			var lit, moved bool
			all := frames(built(t, e, HomeAssistant))
			for i, f := range all {
				if len(f) != Segments {
					t.Fatalf("frame %d has %d segments, want %d", i, len(f), Segments)
				}
				if slices.ContainsFunc(f, func(c Color) bool { return c != Color{} }) {
					lit = true
				}
				if i > 0 && !slices.Equal(f, all[i-1]) {
					moved = true
				}
			}
			if !lit {
				t.Error("never lights a segment")
			}
			if !moved {
				t.Error("never changes")
			}
		})
	}
}

func TestEffectsUseTheColoursTheCatalogueSaysTheyDo(t *testing.T) {
	for _, e := range effects {
		t.Run(e.Name, func(t *testing.T) {
			red, blue := frames(built(t, e, Color{R: 255})), frames(built(t, e, Color{B: 255}))

			if len(e.Palette) > 0 {
				for i := range red {
					if !slices.Equal(red[i], blue[i]) {
						t.Fatalf(
							"frame %d differs with the ring's colour, but this effect has a palette",
							i,
						)
					}
				}
				return
			}

			for i, f := range red {
				for seg, c := range f {
					if c.G != 0 || c.B != 0 {
						t.Fatalf("frame %d segment %d is %+v, want red only", i, seg, c)
					}
				}
			}
		})
	}
}

func TestNamesThatSettingsStoreAreStillOffered(t *testing.T) {
	for _, name := range []string{"Comet", "Rainbow", "Pulse"} {
		if _, ok := byName[name]; !ok {
			t.Errorf("effect %q is gone; settings that stored it now mean nothing", name)
		}
	}
}

func TestEveryEffectIsBuildableAndSaysWhereItBelongs(t *testing.T) {
	for _, e := range effects {
		t.Run(e.Name, func(t *testing.T) {
			switch {
			case e.New == nil && e.Senses == nil:
				t.Error("has no frame function at all")
			case e.New != nil && e.Senses != nil:
				t.Error("has two frame functions; which one runs is then a matter of luck")
			}

			if e.Kinds == 0 {
				t.Error("says nowhere it may be used, so nothing will ever offer it")
			}
			if (e.Senses != nil) != e.Kinds.Has(KindRoom) {
				t.Errorf("kinds %04b and watching the room disagree", e.Kinds)
			}
			if e.Senses != nil && e.Kinds != KindRoom {
				t.Errorf("kinds %04b: watching the room is all one of these can do", e.Kinds)
			}
		})
	}
}

func TestTheListsOfferOnlyWhatFitsThem(t *testing.T) {
	for _, name := range Names(KindLight) {
		if byName[name].Senses != nil {
			t.Errorf("the light offers %q, which cannot run without a room", name)
		}
	}
	for _, name := range Names(KindRoom) {
		if byName[name].Senses == nil {
			t.Errorf("the room reaction offers %q, which does not react to anything", name)
		}
	}

	if _, err := effect(EffectRoomGlow, HomeAssistant, Room{}); err == nil {
		t.Error("building a room effect without a room succeeded, want an error")
	}
	if _, err := effect("Nothing Like This", HomeAssistant, Room{}); err == nil {
		t.Error("building an effect that does not exist succeeded, want an error")
	}
}

func TestEffectNamesAreUnique(t *testing.T) {
	if len(byName) != len(effects) {
		t.Errorf(
			"catalogue has %d effects but %d names; one is shadowing another",
			len(effects),
			len(byName),
		)
	}
}

func TestPaletteWrapsRoundAndClampsAlong(t *testing.T) {
	p := Palette{{R: 255}, {G: 255}, {B: 255}}

	if got, want := p.At(0), p.At(1); got != want {
		t.Errorf("At(0) is %+v and At(1) is %+v, want the circle to close", got, want)
	}
	if got, want := p.At(1.0/3), p[1]; got != want {
		t.Errorf("At(1/3) is %+v, want the second colour %+v", got, want)
	}
	if got, want := p.At(-1.0/3), p[2]; got != want {
		t.Errorf("At(-1/3) is %+v, want the last colour %+v", got, want)
	}
	if mid := p.At(
		1.0 / 6,
	); mid.R < 127 || mid.R > 128 || mid.G < 127 || mid.G > 128 ||
		mid.B != 0 {
		t.Errorf("At(1/6) is %+v, want half of each of the first two", mid)
	}

	if got, want := p.Along(0), p[0]; got != want {
		t.Errorf("Along(0) is %+v, want %+v", got, want)
	}
	if got, want := p.Along(1), p[2]; got != want {
		t.Errorf("Along(1) is %+v, want the last colour %+v, not the first", got, want)
	}
	if got, want := p.Along(2), p[2]; got != want {
		t.Errorf("Along(2) is %+v, want it held at the end %+v", got, want)
	}

	one := Palette{HomeAssistant}
	if got := one.Along(0.7); got != HomeAssistant {
		t.Errorf("Along on a single colour gave %+v, want %+v", got, HomeAssistant)
	}
	if got := one.Nth(-5); got != HomeAssistant {
		t.Errorf("Nth on a single colour gave %+v, want %+v", got, HomeAssistant)
	}
}

func TestHeartbeatThumpsTwiceThenRests(t *testing.T) {
	const period = 1400 * time.Millisecond
	frame := heartbeat(Palette{{R: 255}})

	var peaks int
	var prev, rising = 0.0, false
	for step := range int(period / FrameInterval) {
		f := float64(frame(time.Duration(step) * FrameInterval)[0].R)
		if f > prev {
			rising = true
		} else if rising {
			peaks++
			rising = false
		}
		prev = f
	}
	if peaks != 2 {
		t.Errorf("%d thumps in a period, want 2", peaks)
	}

	if end := frame(period - FrameInterval)[0]; end != (Color{}) {
		t.Errorf("period ends at %+v, want dark", end)
	}
}

func TestChaseKeepsEvenGapsAndStepsWithItsColours(t *testing.T) {
	const step = 110 * time.Millisecond
	p := Palette{{R: 255}, {G: 255}, {B: 255}, {R: 255, G: 255}}
	frame := chase(p)

	var first []int
	for s := range 4 {
		var lit []int
		for i, c := range frame(time.Duration(s)*step + step/2) {
			if c == (Color{}) {
				continue
			}
			if want := p.Nth(len(lit)); c != want {
				t.Errorf("step %d segment %d is %+v, want lamp %d's %+v", s, i, c, len(lit), want)
			}
			lit = append(lit, i)
		}
		if len(lit) != Segments/3 {
			t.Fatalf("step %d lit %v, want %d segments", s, lit, Segments/3)
		}
		for i := 1; i < len(lit); i++ {
			if lit[i]-lit[i-1] != 3 {
				t.Errorf("step %d lit %v, want every third segment", s, lit)
			}
		}

		switch s {
		case 0:
			first = lit
		case 3:
			if !slices.Equal(lit, first) {
				t.Errorf("after a full cycle lit %v, want %v", lit, first)
			}
		default:
			if lit[0] != first[0]+s {
				t.Errorf("step %d starts at %d, want %d", s, lit[0], first[0]+s)
			}
		}
	}
}
