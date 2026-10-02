package settings

import (
	"testing"

	"github.com/ygelfand/echolocal/internal/ui/visual"
)

func TestEveryVisualHasAThumbnail(t *testing.T) {
	for _, k := range visual.Built() {
		if visual.Thumbnail(k, false) == nil {
			t.Errorf("%s has no thumbnail; run make thumbs", k)
		}
	}
}
