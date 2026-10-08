package board

import (
	"log/slog"
	"sync"

	"github.com/ygelfand/echolocal/internal/android/prop"
)

const deviceProp = "ro.product.device"

// Set with -ldflags -X by `make build-echod BOARD=`.
var pinned string

var (
	once      sync.Once
	detected  Board
	askDevice = func() (string, error) { return prop.Get(deviceProp) }
)

func Detect() Board {
	once.Do(func() { detected = detect() })
	return detected
}

func detect() Board {
	said, err := askDevice()
	if err != nil {
		said = ""
	}

	if pinned != "" {
		b, ok := ByCodename(pinned)
		if !ok {
			slog.Error(
				"built for a board this binary has no entry for, assuming biscuit",
				"pinned",
				pinned,
			)
			return Biscuit
		}
		if said != "" && said != b.Device {
			slog.Warn(
				"built for one board and running on another",
				"pinned",
				b.Codename,
				"device",
				said,
			)
		}
		slog.Info("board", "codename", b.Codename, "pinned", true)
		return b
	}

	if said == "" {
		slog.Warn("reading the board failed, assuming biscuit", "prop", deviceProp, "err", err)
		return Biscuit
	}

	b, ok := For(said)
	if !ok {
		slog.Warn("unknown board, assuming biscuit", "device", said)
		return Biscuit
	}
	slog.Info("board", "codename", b.Codename, "device", b.Device)
	return b
}
