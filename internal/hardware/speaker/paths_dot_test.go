//go:build !board_checkers && !board_cronos && !board_crown && !board_athena

package speaker

import (
	"slices"
	"testing"
)

func TestEveryDotPathRoutesTheDAC(t *testing.T) {
	for out, seq := range pathSequence {
		for _, name := range []string{"HPL Output Mixer L_DAC Switch", "HPR Output Mixer R_DAC Switch"} {
			on := func(k kctl) bool { return k.name == name && k.level == 1 }
			if !slices.ContainsFunc(seq, on) {
				t.Errorf("the %s path leaves %q off", out, name)
			}
		}
	}
}
