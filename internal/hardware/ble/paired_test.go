//go:build linux && (board_checkers || board_cronos || board_rook)

package ble

import (
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

// paired is a port whose writes go somewhere that never replies, and the other end to read them
// off. A socketpair rather than a pipe because Send writes and a reader reads the same port.
func paired(t *testing.T) (*Port, *os.File) {
	t.Helper()

	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("socketpair: %v", err)
	}

	// Non-blocking, or os.NewFile hands back a file the runtime cannot poll and every deadline
	// the port sets fails outright — which looks like a timeout without being one.
	for _, fd := range fds {
		if err := unix.SetNonblock(fd, true); err != nil {
			t.Fatalf("nonblock: %v", err)
		}
	}

	mine := os.NewFile(uintptr(fds[0]), "port")
	theirs := os.NewFile(uintptr(fds[1]), "chip")
	t.Cleanup(func() { mine.Close(); theirs.Close() })

	return &Port{f: mine}, theirs
}
