package led

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func ringInDir(t *testing.T) *Ring {
	t.Helper()

	dir := t.TempDir()
	for _, name := range []string{"frame", "leds_current", "boot_animation"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatalf("preparing %s: %v", name, err)
		}
	}
	return &Ring{Path: dir}
}

func shown(t *testing.T, r *Ring) Color {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for {
		r.mu.Lock()
		drawn := len(r.last) >= 3
		var c Color
		if drawn {
			c = Color{R: r.last[0], G: r.last[1], B: r.last[2]}
		}
		r.mu.Unlock()

		if drawn {
			return c
		}
		if time.Now().After(deadline) {
			t.Fatal("the ring was never written")
		}
		time.Sleep(2 * time.Millisecond)
	}
}

func settle() { time.Sleep(40 * time.Millisecond) }

func waitFor(t *testing.T, r *Ring, want Color) Color {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	var got Color
	for time.Now().Before(deadline) {
		if got = shown(t, r); got == want {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	return got
}

func solid(c Color) []Color {
	out := make([]Color, Segments)
	for i := range out {
		out[i] = c
	}
	return out
}

var (
	red   = Color{R: 0xC0}
	green = Color{G: 0xC0}
	blue  = Color{B: 0xC0}
)

func running(t *testing.T) *Driver {
	t.Helper()

	d := NewDriver(ringInDir(t))
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	go func() { _ = d.Run(ctx) }()
	return d
}

func TestPriorityWinsAndUncovers(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))
	settle()
	if got := shown(t, d.ring); got != green {
		t.Fatalf("base alone shows %+v, want green", got)
	}

	turn := d.Claim(PriorityTurn)
	turn.Paint(solid(blue))
	settle()
	if got := shown(t, d.ring); got != blue {
		t.Fatalf("with a turn claim the ring shows %+v, want blue", got)
	}

	trouble := d.Claim(PriorityTrouble)
	trouble.Paint(solid(red))
	settle()
	if got := shown(t, d.ring); got != red {
		t.Fatalf("with trouble up the ring shows %+v, want red", got)
	}

	turn.Release()
	settle()
	if got := shown(t, d.ring); got != red {
		t.Fatalf("after the turn ended the ring shows %+v, want red still", got)
	}

	trouble.Release()
	settle()
	if got := shown(t, d.ring); got != green {
		t.Fatalf("after trouble released the ring shows %+v, want the base back", got)
	}
}

func TestEmptyClaimShowsNothing(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))
	settle()

	_ = d.Claim(PriorityBoot)
	settle()

	if got := shown(t, d.ring); got != green {
		t.Errorf("an empty claim changed the ring to %+v", got)
	}
}

func TestTimedClaimExpires(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))

	notice := d.Claim(PriorityNotice)
	notice.PaintFor(solid(blue), 80*time.Millisecond)
	settle()
	if got := shown(t, d.ring); got != blue {
		t.Fatalf("the notice shows %+v, want blue", got)
	}

	time.Sleep(140 * time.Millisecond)
	if got := shown(t, d.ring); got != green {
		t.Errorf("after the notice expired the ring shows %+v, want the base back", got)
	}
}

func TestLatestOfEqualPriorityWins(t *testing.T) {
	d := running(t)

	first := d.Claim(PriorityNotice)
	first.Paint(solid(green))
	second := d.Claim(PriorityNotice)
	second.Paint(solid(blue))
	settle()

	if got := shown(t, d.ring); got != blue {
		t.Errorf("the ring shows %+v, want the later claim", got)
	}

	second.Release()
	settle()
	if got := shown(t, d.ring); got != green {
		t.Errorf("after the later claim released the ring shows %+v, want the earlier one", got)
	}
}

func TestContentSwap(t *testing.T) {
	d := running(t)

	c := d.Claim(PriorityTurn)
	c.Paint(solid(green))
	settle()

	c.Paint(solid(red))
	settle()
	if got := shown(t, d.ring); got != red {
		t.Errorf("after a swap the ring shows %+v, want red", got)
	}

	c.Clear()
	settle()
	if got := shown(t, d.ring); got != (Color{}) {
		t.Errorf("after clearing the only claim the ring shows %+v, want blank", got)
	}
}

func TestLowerClaimDoesNotRestartTheTop(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))

	var starts atomic.Int32
	boot := d.Claim(PriorityBoot)
	boot.Show(Content{Animate: func(ctx context.Context, r *Ring) error {
		starts.Add(1)
		if err := r.SetSegments(solid(blue)); err != nil {
			return err
		}
		<-ctx.Done()
		return nil
	}})
	settle()

	if n := starts.Load(); n != 1 {
		t.Fatalf("the animation started %d times before anything changed", n)
	}

	for range 3 {
		base.Paint(solid(red))
		settle()
		base.Paint(solid(green))
		settle()
	}

	if n := starts.Load(); n != 1 {
		t.Errorf("the animation restarted: %d starts", n)
	}
	if got := shown(t, d.ring); got != blue {
		t.Errorf("the ring shows %+v, want the boot animation still", got)
	}

	boot.Show(Content{Frame: solid(red)})
	settle()
	if got := shown(t, d.ring); got != red {
		t.Errorf("after changing the top claim the ring shows %+v, want red", got)
	}
}

func TestReleasedClaimIsInert(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))

	gone := d.Claim(PriorityTrouble)
	gone.Paint(solid(red))
	settle()
	gone.Release()
	settle()

	gone.Paint(solid(blue))
	settle()
	if got := shown(t, d.ring); got != green {
		t.Errorf("a released claim painted %+v", got)
	}
}

func TestARoomReactionShowsTheLightThroughItsSilence(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))

	var loud atomic.Bool
	room := d.Claim(PriorityRoom)
	room.React(EffectRoomGlow, blue, Room{Level: func() float64 {
		if loud.Load() {
			return 1
		}
		return 0
	}})
	settle()

	if got := shown(t, d.ring); got != green {
		t.Errorf("a silent room shows %+v, want the light underneath", got)
	}

	loud.Store(true)
	if got := waitFor(t, d.ring, blue); got != blue {
		t.Errorf("a loud room shows %+v, want the reaction", got)
	}

	loud.Store(false)
	if got := waitFor(t, d.ring, green); got != green {
		t.Errorf("once the room went quiet the ring shows %+v, want the light again", got)
	}
}

func TestRepaintingWhatShowsThroughReachesTheRing(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))

	room := d.Claim(PriorityRoom)
	room.React(EffectRoomGlow, blue, Room{Level: func() float64 { return 0 }})
	settle()

	if got := shown(t, d.ring); got != green {
		t.Fatalf("a silent room shows %+v, want the light underneath", got)
	}

	base.Paint(solid(red))
	if got := waitFor(t, d.ring, red); got != red {
		t.Errorf("after repainting underneath the ring shows %+v, want the new colour", got)
	}
}

func TestADarkAnimationStillOwnsTheRing(t *testing.T) {
	d := running(t)

	base := d.Claim(PriorityBase)
	base.Paint(solid(green))

	turn := d.Claim(PriorityTurn)
	turn.Play(EffectHeartbeat, red)

	for shown(t, d.ring) == green {
		time.Sleep(FrameInterval)
	}

	for range 30 {
		if got := shown(t, d.ring); got == green {
			t.Fatal("the light underneath showed through a resting heartbeat")
		}
		time.Sleep(FrameInterval)
	}
}
