//go:build board_checkers || board_cronos || board_rook

package settings

import (
	"testing"

	sharedvisual "github.com/ygelfand/libcountertop/pkg/display/visual"

	"github.com/ygelfand/echolocal/internal/ui/visual"
)

func TestEveryVisualHasAThumbnail(t *testing.T) {
	for _, k := range sharedvisual.Built() {
		if visual.Thumbs().Thumbnail(k, false) == nil {
			t.Errorf("%s has no thumbnail; run make thumbs", k)
		}
	}
}
