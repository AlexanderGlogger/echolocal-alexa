package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func released(from string) Manifest {
	b := Binary{
		URL:    from + "/releases/download/0.11.1/echod-arm64",
		SHA256: strings.Repeat("a", 64),
		Size:   1,
	}
	return Manifest{Version: "0.11.1", Binaries: map[string]Binary{"arm64": b}}
}

func TestSelfUpdateRefusedOnlyWithTheBridgeInstalled(t *testing.T) {
	dir := t.TempDir()
	jar := filepath.Join(dir, "amazon-helper.jar")
	upstream := released("https://github.com/ygelfand/echolocal")

	if err := refusePryonAt(jar, upstream); err != nil {
		t.Fatalf("refused an update with no bridge installed: %v", err)
	}

	// A directory where the jar belongs is not a bridge, and the bridge is not enabled for it either.
	if err := os.Mkdir(jar, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := refusePryonAt(jar, upstream); err != nil {
		t.Fatalf("refused an update over a directory: %v", err)
	}
	if err := os.Remove(jar); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(jar, []byte("dex"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := refusePryonAt(jar, upstream); !errors.Is(err, errPryon) {
		t.Fatalf("got %v with the bridge installed, want the Pryon refusal", err)
	}
}

func TestBridgeTakesOnlyTheForksOwnReleases(t *testing.T) {
	jar := filepath.Join(t.TempDir(), "amazon-helper.jar")
	if err := os.WriteFile(jar, []byte("dex"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := refusePryonAt(jar, released(forkRepo)); err != nil {
		t.Fatalf("refused a release from %s: %v", forkRepo, err)
	}

	mixed := released(forkRepo)
	mixed.Boards = map[string]map[string]Binary{
		"crown": {"arm64": released("https://github.com/ygelfand/echolocal").Binaries["arm64"]},
	}
	if err := refusePryonAt(jar, mixed); !errors.Is(err, errPryon) {
		t.Fatalf("got %v for a manifest naming an upstream binary, want the Pryon refusal", err)
	}

	flat := Manifest{
		Version: "0.11.1",
		URL:     "https://example.com/echod",
		SHA256:  strings.Repeat("a", 64),
		Size:    1,
	}
	if err := refusePryonAt(jar, flat); !errors.Is(err, errPryon) {
		t.Fatalf("got %v for a flat manifest from elsewhere, want the Pryon refusal", err)
	}
}

// A resync that takes upstream's side of channel.go or the Makefile would quietly point the Dots, or the
// manifests CI writes, back at upstream's releases.
func TestChannelsAndReleasesPointAtTheFork(t *testing.T) {
	if repo != forkRepo {
		t.Errorf("the channels fetch from %s, want %s", repo, forkRepo)
	}

	makefile, err := os.ReadFile(filepath.Join("..", "..", "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	want := "REPO ?= " + strings.TrimPrefix(forkRepo, "https://github.com/")
	for line := range strings.SplitSeq(string(makefile), "\n") {
		if strings.HasPrefix(line, "REPO ?=") {
			if strings.TrimSpace(line) != want {
				t.Errorf("the Makefile has %q, want %q", line, want)
			}
			return
		}
	}
	t.Errorf("the Makefile sets no REPO, want %q", want)
}
