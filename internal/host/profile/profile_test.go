package profile

import "testing"

// Fire OS 6 on biscuit has no package manager.
func TestBiscuitHidesNothing(t *testing.T) {
	if len(biscuit.Hide) != 0 {
		t.Errorf("biscuit hides %d packages: %v", len(biscuit.Hide), biscuit.Hide)
	}
}

func TestEveryProfileNamesAService(t *testing.T) {
	for device, p := range profiles {
		if p.Board.ServiceName == "" || p.Board.Service == "" {
			t.Errorf("%s names no service for echod to take over", device)
		}
		if _, err := For(device); err != nil {
			t.Errorf("%s has a complete profile and was still refused: %v", device, err)
		}
	}

	if _, err := For("sailfish"); err == nil {
		t.Error("a device nobody has heard of resolved a profile")
	}
}

func TestHiddenListsAreDistinct(t *testing.T) {
	for device, p := range profiles {
		seen := make(map[string]bool, len(p.Hide))
		for _, pkg := range p.Hide {
			if seen[pkg.Name] {
				t.Errorf("%s hides %s twice", device, pkg.Name)
			}
			seen[pkg.Name] = true

			if pkg.Reason == "" {
				t.Errorf("%s hides %s with no reason", device, pkg.Name)
			}
		}
	}
}
