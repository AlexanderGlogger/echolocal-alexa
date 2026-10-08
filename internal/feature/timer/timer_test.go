package timer

import (
	"testing"
	"time"

	"github.com/ygelfand/echolocal/internal/hardware/led"
)

func lit(frame []led.Color) int {
	n := 0
	for _, c := range frame {
		if c != (led.Color{}) {
			n++
		}
	}
	return n
}

func TestTheRingEmptiesAsATimerRunsDown(t *testing.T) {
	total := time.Minute
	for _, c := range []struct {
		left time.Duration
		want int
	}{
		{time.Minute, 12},
		{30 * time.Second, 6},
		{5 * time.Second, 1},
		{0, 0},
	} {
		if got := lit(frame(c.left, total)); got != c.want {
			t.Errorf("%s left of %s: %d segments, want %d", c.left, total, got, c.want)
		}
	}
}

func TestTheLeadingSegmentCarriesThePartOfItStillToRun(t *testing.T) {
	f := frame(32*time.Second, time.Minute)

	if f[6] == (led.Color{}) || f[6] == countdownColor {
		t.Errorf(
			"the leading segment is %v, want it dimmed between black and %v",
			f[6],
			countdownColor,
		)
	}
	if f[5] != countdownColor {
		t.Errorf("the segment behind it is %v, want %v", f[5], countdownColor)
	}
	if f[7] != (led.Color{}) {
		t.Errorf("the segment ahead of it is %v, want black", f[7])
	}
}

func TestATimerWithNoTotalDoesNotDivideByIt(t *testing.T) {
	if got := lit(frame(time.Second, 0)); got != 12 {
		t.Errorf("%d segments, want 12", got)
	}
}
