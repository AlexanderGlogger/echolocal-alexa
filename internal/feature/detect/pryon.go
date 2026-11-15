package detect

import (
	"log/slog"
	"slices"

	"github.com/ygelfand/libcountertop/pkg/inference/detect"
	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/lib/wake"
)

// choose is detect.Choose with the companion's word kept out of the engine. Amazon's detector scores
// in its own Android process and reports only a wake, so its slot is accepted with nothing loaded,
// and nothing waits for it to score. Upstream's own Pryon words go to the engine like any other.
func choose(
	e *detect.Engine,
	models []sharedwake.Model,
	ids []string,
	count int,
) (accepted []string, loaded []int) {
	accepted, loaded = detect.Choose(external{e}, models, ids, count)
	loaded = slices.DeleteFunc(loaded, func(slot int) bool {
		m, _ := sharedwake.Find(models, accepted[slot])
		return wake.External(m)
	})
	return accepted, loaded
}

// external is the engine's slots with the companion's left empty. Keeping that slot out of the
// engine is the boundary that prevents a fake model path or a second detector.
type external struct{ *detect.Engine }

func (e external) Use(slot int, m sharedwake.Model) error {
	if !wake.External(m) {
		return e.Engine.Use(slot, m)
	}
	e.Clear(slot)
	slog.Info("external wake word selected", "slot", slot+1, "id", m.ID, "engine", m.Kind)
	return nil
}
