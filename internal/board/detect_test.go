package board

import "testing"

func withDevice(t *testing.T, said string, err error) {
	t.Helper()
	restore := askDevice
	askDevice = func() (string, error) { return said, err }
	t.Cleanup(func() { askDevice = restore })
}

func withPin(t *testing.T, name string) {
	t.Helper()
	restore := pinned
	pinned = name
	t.Cleanup(func() { pinned = restore })
}

func TestDetectReadsTheDevice(t *testing.T) {
	withDevice(t, "checkers", nil)

	if got := detect(); got.Codename != "checkers" {
		t.Errorf("detect() = %s, want checkers", got)
	}
}

func TestAPinnedBuildDoesNotAskTheDevice(t *testing.T) {
	withPin(t, "checkers")
	withDevice(t, "biscuit_puffin", nil)

	if got := detect(); got.Codename != "checkers" {
		t.Errorf("detect() = %s, want the pinned checkers", got)
	}
}

func TestAPinNothingMatchesFallsBackToBiscuit(t *testing.T) {
	withPin(t, "nonesuch")
	withDevice(t, "", errNoProp)

	if got := detect(); got.Codename != Biscuit.Codename {
		t.Errorf("detect() = %s, want biscuit", got)
	}
}

func TestDetectFallsBackToBiscuit(t *testing.T) {
	for _, tc := range []struct {
		name, said string
		err        error
	}{
		{"prop unreadable", "", errNoProp},
		{"unknown device", "sailfish", nil},
		{"no device recorded for it", "crown", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withDevice(t, tc.said, tc.err)

			if got := detect(); got.Codename != Biscuit.Codename {
				t.Errorf("detect() = %s, want biscuit", got)
			}
		})
	}
}

var errNoProp = errNotThere{}

type errNotThere struct{}

func (errNotThere) Error() string { return "no such file" }
