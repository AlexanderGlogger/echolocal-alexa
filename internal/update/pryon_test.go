package update

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSelfUpdateRefusedOnlyWithTheBridgeInstalled(t *testing.T) {
	dir := t.TempDir()
	jar := filepath.Join(dir, "amazon-helper.jar")

	if err := refusePryonAt(jar); err != nil {
		t.Fatalf("refused an update with no bridge installed: %v", err)
	}

	// A directory where the jar belongs is not a bridge, and the bridge is not enabled for it either.
	if err := os.Mkdir(jar, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := refusePryonAt(jar); err != nil {
		t.Fatalf("refused an update over a directory: %v", err)
	}
	if err := os.Remove(jar); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(jar, []byte("dex"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := refusePryonAt(jar); !errors.Is(err, errPryon) {
		t.Fatalf("got %v with the bridge installed, want the Pryon refusal", err)
	}
}
