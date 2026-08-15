package detect

import (
	"log/slog"
	"slices"

	"github.com/ygelfand/libcountertop/pkg/inference/detect"
	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/lib/wake"
)

// choose is detect.Choose with Pryon kept out of the PCM engine. Amazon's detector scores in its own
// Android process and reports only a wake, so its slot is accepted with nothing loaded, and nothing
// waits for it to score.
func choose(
	e *detect.Engine,
	models []sharedwake.Model,
	ids []string,
	count int,
) (accepted []string, loaded []int) {
	accepted, loaded = detect.Choose(external{e}, models, ids, count)
	loaded = slices.DeleteFunc(loaded, func(slot int) bool {
		m, _ := sharedwake.Find(models, accepted[slot])
		return m.Kind == wake.KindPryon
	})
	return accepted, loaded
}

// external is the engine's slots with Pryon's left empty. Keeping that slot out of the engine is
// the boundary that prevents a fake model path or a second detector.
type external struct{ *detect.Engine }

func (e external) Use(slot int, m sharedwake.Model) error {
	if m.Kind != wake.KindPryon {
		return e.Engine.Use(slot, m)
	}
	e.Clear(slot)
	slog.Info("external wake word selected", "slot", slot+1, "id", m.ID, "engine", m.Kind)
	return nil
}
