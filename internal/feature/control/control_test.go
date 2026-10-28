package control

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"
)

// sockaddr_un holds 104 bytes on darwin.
func sock(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "c")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return filepath.Join(dir, "s")
}

func connect(t *testing.T) net.Conn {
	t.Helper()

	c := build(sock(t))
	if err := c.Start(); err != nil {
		t.Fatalf("up: %v", err)
	}
	t.Cleanup(c.Close)

	conn, err := net.DialTimeout("unix", c.Address(), 2*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatal(err)
	}
	return conn
}

func say(t *testing.T, conn net.Conn, in *bufio.Scanner, line string) (string, bool) {
	t.Helper()

	if _, err := fmt.Fprintln(conn, line); err != nil {
		t.Fatalf("send %q: %v", line, err)
	}

	var out strings.Builder
	if err := harness.Answer(in, &out); err != nil {
		return "error: " + err.Error(), false
	}
	return out.String(), true
}

func TestACommandIsAnsweredAndTerminated(t *testing.T) {
	conn := connect(t)
	in := bufio.NewScanner(conn)

	out, ok := say(t, conn, in, "device version")
	if !ok {
		t.Fatalf("version failed: %s", out)
	}
	if !strings.Contains(out, "board") {
		t.Errorf("version said %q, want it to name the board", out)
	}
}

func TestAnUnknownCommandIsAnError(t *testing.T) {
	conn := connect(t)
	in := bufio.NewScanner(conn)

	out, ok := say(t, conn, in, "nonesuch")
	if ok {
		t.Fatalf("an unknown command reported success: %s", out)
	}
	if !strings.HasPrefix(out, "error: ") {
		t.Errorf("said %q, want an error: line", out)
	}
}

func TestFirmwareCommandsAreTopLevel(t *testing.T) {
	conn := connect(t)
	in := bufio.NewScanner(conn)
	for _, command := range []string{"firmware --help", "firmware list --help", "firmware install --help"} {
		out, ok := say(t, conn, in, command)
		if !ok || !strings.Contains(out, "firmware") {
			t.Fatalf("%s: %q", command, out)
		}
	}
}

func TestABlankLineIsAnswered(t *testing.T) {
	conn := connect(t)
	in := bufio.NewScanner(conn)

	if _, ok := say(t, conn, in, ""); !ok {
		t.Error("a blank line was reported as a failure")
	}
}

func TestOneConnectionTakesSeveralCommands(t *testing.T) {
	conn := connect(t)
	in := bufio.NewScanner(conn)

	for _, cmd := range []string{"device version", "device state", "device version"} {
		if out, ok := say(t, conn, in, cmd); !ok {
			t.Fatalf("%s failed: %s", cmd, out)
		}
	}
}

func TestDownReleasesTheAddress(t *testing.T) {
	c := build(sock(t))
	if err := c.Start(); err != nil {
		t.Fatalf("up: %v", err)
	}
	c.Close()

	if err := c.Start(); err != nil {
		t.Fatalf("reopening after down: %v", err)
	}
	c.Close()
}
