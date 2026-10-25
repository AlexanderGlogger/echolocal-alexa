//go:build !board_checkers && !board_cronos && !board_crown && !board_athena

package speaker

import (
	"slices"
	"testing"
)

var dacSwitches = []string{"HPL Output Mixer L_DAC Switch", "HPR Output Mixer R_DAC Switch"}

func TestTheDotRoutesTheDACAtStartup(t *testing.T) {
	for _, name := range dacSwitches {
		on := func(k kctl) bool { return k.name == name && k.level == 1 }
		if !slices.ContainsFunc(initSequence, on) {
			t.Errorf("the init sequence leaves %q off", name)
		}
	}
}

func TestNoDotSequenceTurnsTheDACOff(t *testing.T) {
	seqs := sequencesFor(nil)
	all := map[string][]kctl{"on": seqs.on, "off": seqs.off, "close": seqs.close}
	for out, seq := range seqs.path {
		all["path "+string(out)] = seq
	}
	for out, seq := range seqs.leave {
		all["leave "+string(out)] = seq
	}
	for which, seq := range all {
		for _, k := range seq {
			if slices.Contains(dacSwitches, k.name) && k.level != 1 {
				t.Errorf("the %s sequence sets %q to %d", which, k.name, k.level)
			}
		}
	}
}
