package detect

import (
	"slices"
	"testing"

	"github.com/ygelfand/libcountertop/pkg/inference/detect"
	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/lib/wake"
)

// Selecting Pryon has to be accepted without loading anything: the engine would otherwise build a
// microWakeWord detector from an empty path, and the busy light would wait on a slot that never
// scores.
func TestChooseAcceptsPryonWithoutLoadingIt(t *testing.T) {
	e := detect.New(2, nil)
	models := []sharedwake.Model{wake.PryonModel()}

	accepted, loaded := choose(e, models, []string{"", wake.PryonID}, 2)

	if !slices.Equal(accepted, []string{"", wake.PryonID}) {
		t.Errorf("accepted %v, want Pryon in the second slot", accepted)
	}
	if len(loaded) != 0 {
		t.Errorf("waiting on slots %v to score, want none", loaded)
	}
	if _, ok := e.Loaded(1); ok {
		t.Error("Pryon was loaded into the PCM engine")
	}
}
