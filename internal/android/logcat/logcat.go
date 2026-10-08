package logcat

import (
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"

	"github.com/ygelfand/echolocal/internal/android/logd"
)

const socket = "/dev/socket/logdr"

const (
	lidMain   = 0
	lidSystem = 3
)

// The oldest logger_entry layout has a 20-byte header; later ones carry hdr_size.
const hdrSize = 20

type Entry struct {
	Level slog.Level
	Tag   string
	Text  string
}

func Recent(lines int) ([]Entry, error) {
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_SEQPACKET, 0)
	if err != nil {
		return nil, fmt.Errorf("logcat: socket: %w", err)
	}

	f := os.NewFile(uintptr(fd), socket)
	defer f.Close()

	if err := syscall.Connect(fd, &syscall.SockaddrUnix{Name: socket}); err != nil {
		return nil, fmt.Errorf("logcat: connect %s: %w", socket, err)
	}

	req := fmt.Sprintf("dumpAndClose lids=%d,%d tail=%d", lidMain, lidSystem, lines)
	if _, err := f.Write([]byte(req)); err != nil {
		return nil, fmt.Errorf("logcat: %q: %w", req, err)
	}

	var out []Entry
	buf := make([]byte, 5*1024)
	for {
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			return out, nil
		}
		if e, ok := parse(buf[:n]); ok {
			out = append(out, e)
		}
	}
}

func parse(pkt []byte) (Entry, bool) {
	if len(pkt) < hdrSize {
		return Entry{}, false
	}

	start := int(binary.LittleEndian.Uint16(pkt[2:]))
	if start < hdrSize || start > len(pkt) {
		start = hdrSize
	}

	body := pkt[start:]
	if len(body) < 3 {
		return Entry{}, false
	}

	parts := strings.SplitN(strings.TrimRight(string(body[1:]), "\x00"), "\x00", 2)
	if len(parts) < 2 {
		return Entry{}, false
	}

	return Entry{Level: logd.Level(body[0]), Tag: parts[0], Text: parts[1]}, true
}
