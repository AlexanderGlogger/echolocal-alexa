package led

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetFrameSkipsIdenticalWrites(t *testing.T) {
	dir := t.TempDir()
	r := &Ring{Path: dir}
	frame := filepath.Join(dir, "frame")

	first := make([]byte, Channels)
	first[0] = 0x10
	if err := r.SetFrame(first); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(frame); err != nil {
		t.Fatal(err)
	}
	if err := r.SetFrame(first); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(frame); !os.IsNotExist(err) {
		t.Fatal("an identical frame was written again")
	}

	changed := make([]byte, Channels)
	changed[0] = 0x11
	if err := r.SetFrame(changed); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(frame); err != nil {
		t.Fatal("a changed frame was not written:", err)
	}

	changed[0] = 0x12
	if err := os.Remove(frame); err != nil {
		t.Fatal(err)
	}
	if err := r.SetFrame(changed); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(frame); err != nil {
		t.Fatal("a frame differing only from a reused buffer was skipped:", err)
	}
}

func TestForgetForcesTheNextWrite(t *testing.T) {
	dir := t.TempDir()
	r := &Ring{Path: dir}
	frame := filepath.Join(dir, "frame")

	vals := make([]byte, Channels)
	if err := r.SetFrame(vals); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(frame); err != nil {
		t.Fatal(err)
	}

	r.Forget()
	if err := r.SetFrame(vals); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(frame); err != nil {
		t.Fatal("Forget did not force the next write:", err)
	}
}
