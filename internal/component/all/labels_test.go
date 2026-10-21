package all

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ygelfand/libcountertop/pkg/runtime/webapi"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
)

func TestEverySettingHasItsLabels(t *testing.T) {
	config.Use(filepath.Join(t.TempDir(), "state.json"))
	component.Default().Entities()
	if missing := component.Settings.Unlabeled(); len(missing) > 0 {
		t.Errorf("en.yaml has no text for:\n\t%s", strings.Join(missing, "\n\t"))
	}
	if loose := webapi.Unplaced(&component.Settings); len(loose) > 0 {
		t.Errorf("webapi sections hold no place for:\n\t%s", strings.Join(loose, "\n\t"))
	}
}
