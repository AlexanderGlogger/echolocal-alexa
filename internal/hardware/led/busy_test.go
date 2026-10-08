package led

import "testing"

func busyFixture() *Busy { return NewDriver(nil).Busy() }

func TestBusyShowsWorkUntilItIsDone(t *testing.T) {
	b := busyFixture()

	if got := b.claim.get(); !got.empty() {
		t.Fatalf("the ring is showing %+v before any work started", got)
	}

	task := b.Start(WorkWakeWord)
	if !b.Running() {
		t.Error("not running with work started")
	}

	effect, base := WorkWakeWord.appearance()
	if got := b.claim.get(); got.Effect != effect || got.Base != base {
		t.Errorf("showing %q in %+v, want %q in %+v", got.Effect, got.Base, effect, base)
	}

	task.Done()
	if b.Running() {
		t.Error("still running after the work was done")
	}
	if got := b.claim.get(); !got.empty() {
		t.Errorf("the ring is still showing %+v", got)
	}
}

func TestBusyHoldsUntilTheLastWorkFinishes(t *testing.T) {
	b := busyFixture()

	first, second := b.Start(WorkWakeWord), b.Start(WorkWakeWord)

	second.Done()
	if !b.Running() {
		t.Fatal("gave the ring back with work still outstanding")
	}

	first.Done()
	if b.Running() {
		t.Error("still running with nothing outstanding")
	}
}

func TestBusyDoneTwiceLeavesOtherWorkAlone(t *testing.T) {
	b := busyFixture()

	task := b.Start(WorkWakeWord)
	task.Done()

	b.Start(WorkWakeWord)
	task.Done()

	if !b.Running() {
		t.Error("a second Done on finished work gave away the ring")
	}
}

func TestBusyDoneOnNothing(t *testing.T) {
	var task *Task
	task.Done()
}
