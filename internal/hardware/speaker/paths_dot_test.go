//go:build !board_checkers && !board_cronos && !board_crown && !board_athena

package speaker

import (
	"slices"
	"testing"
)

func TestEveryDotPathRoutesTheDAC(t *testing.T) {
	for out, seq := range pathSequence {
		for _, name := range []string{"HPL Output Mixer L_DAC Switch", "HPR Output Mixer R_DAC Switch"} {
			if !slices.ContainsFunc(seq, func(k kctl) bool { return k.name == name && k.level == 1 }) {
				t.Errorf("the %s path leaves %q off", out, name)
			}
		}
	}
}
