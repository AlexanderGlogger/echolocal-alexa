package installer

import (
	"strings"
	"testing"

	"github.com/ygelfand/echolocal/internal/host/bootimg"
)

func TestReadyNeedsRootAndPermissiveOutsideRecovery(t *testing.T) {
	for _, tc := range []struct {
		rooted, permissive, recovery, want bool
	}{
		{true, true, false, true},
		{true, true, true, false},
		{true, false, false, false},
		{false, true, false, false},
		{false, false, false, false},
		{false, false, true, false},
	} {
		s := state{rooted: tc.rooted, permissive: tc.permissive, recovery: tc.recovery}
		if got := s.ready(); got != tc.want {
			t.Errorf("ready(root=%t permissive=%t recovery=%t) = %t, want %t",
				tc.rooted, tc.permissive, tc.recovery, got, tc.want)
		}
	}
}

func TestStateSaysWhatItFound(t *testing.T) {
	if said := (state{rooted: false}).String(); !strings.Contains(said, "root=false") {
		t.Errorf("%q does not say whether it found root", said)
	}
}

func TestWritingStepsSkipWhenReady(t *testing.T) {
	r := &run{state: state{rooted: true, permissive: true}}

	detail, skip := r.done()
	if !skip {
		t.Fatal("a ready device is not skipping the writing steps")
	}
	if detail == "" {
		t.Error("skipped with no reason given")
	}

	r.state = state{rooted: false}
	if _, skip := r.done(); skip {
		t.Error("a device without root is skipping the write")
	}
}

func TestNothingIsWrittenWithoutApproval(t *testing.T) {
	r := &run{cfg: Config{}, state: state{rooted: false}}

	if _, _, err := checkApproval(r); err == nil {
		t.Fatal("unapproved write was allowed")
	}

	r.cfg.Approved = true
	if _, _, err := checkApproval(r); err != nil {
		t.Errorf("approved write refused: %v", err)
	}

	r.cfg.Approved = false
	r.state = state{rooted: true, permissive: true}
	_, skip, err := checkApproval(r)
	if err != nil || !skip {
		t.Errorf("ready device: skip=%t err=%v, want skipped with no error", skip, err)
	}
}

func TestCheckImageRefusesAnEmptyOne(t *testing.T) {
	flashes := &run{state: state{device: "biscuit", build: "272.6.8.0_user_680767620"}}
	flashes.cfg.Profile.Boot = bootimg.Biscuit

	if _, _, err := checkImage(flashes); err == nil {
		t.Error("accepted an empty image")
	}

	keeps := &run{state: state{device: "rook"}}
	if _, skip, err := checkImage(keeps); err != nil || !skip {
		t.Errorf(
			"board with no image of its own: skip=%t err=%v, want skipped with no error",
			skip,
			err,
		)
	}
}
