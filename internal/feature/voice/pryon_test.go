package voice

import (
	"testing"

	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"
	"github.com/ygelfand/libcountertop/pkg/inference/wakeslots"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/lib/wake"
)

func TestPryonInstallPreselectsAlexa(t *testing.T) {
	config.Use(t.TempDir() + "/state.json")
	models := []sharedwake.Model{
		{ID: sharedwake.DefaultModel, Phrase: "Okay Nabu"},
		wake.PryonModel(),
	}

	active := wakeslots.Advertised(wanted(models, wakeword.Slots))
	if len(active) != 1 || active[0] != wake.PryonID {
		t.Errorf("listening for %v, want just %q", active, wake.PryonID)
	}
}
