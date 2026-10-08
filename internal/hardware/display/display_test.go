package display

import (
	"testing"

	"github.com/ygelfand/libcountertop/pkg/display/geometry"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
)

func TestTheShowsAreMountedLandscapeInverted(t *testing.T) {
	was := component.Board()
	defer component.Default().Use(was)

	for _, b := range []board.Board{board.Checkers, board.Cronos} {
		component.Default().Use(b)
		if Mounted() != geometry.Rotate270 {
			t.Errorf("%s is mounted %v, want %v", b.Codename, Mounted(), geometry.Rotate270)
		}
	}
}
