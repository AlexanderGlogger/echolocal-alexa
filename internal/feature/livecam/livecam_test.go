package livecam

import (
	"strings"
	"testing"

	"github.com/ygelfand/libcountertop/pkg/camera/live"
)

func TestEveryCameraNameIsTranslated(t *testing.T) {
	table := live.New(live.Options{
		Vendor:  vendor,
		Streams: live.Streams{Main: []string{"1280x720"}},
	}).Table()
	for _, g := range table.Groups() {
		if got := table.Title(g); strings.HasPrefix(got, "camera.") {
			t.Errorf("the %s section has no text: %q", g, got)
		}
	}
	for _, s := range table.Rows() {
		if got := s.Title(); strings.HasPrefix(got, "camera.") {
			t.Errorf("%s has no text: %q", s.Name, got)
		}
		for _, o := range s.Options {
			if got := s.Label(o); got == "" || strings.HasPrefix(got, "camera.") {
				t.Errorf("%s option %q has no text: %q", s.Name, o.Value, got)
			}
		}
	}
}
