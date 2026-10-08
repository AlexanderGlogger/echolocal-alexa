package led

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/ygelfand/echolocal/internal/hardware/indicate"
)

// DefaultPath is the is31fl3236 i2c device directory.
const DefaultPath = "/sys/bus/i2c/devices/0-003f"

const Channels = 36

const Segments = 12

// Segment n is channels 3n..3n+2 in RGB order; segment 0 sits between mic and volume-down, indices run clockwise.

type Color = indicate.Color

func (r *Ring) SetSegments(c []Color) error {
	if len(c) != Segments {
		return fmt.Errorf("led: need %d segments, got %d", Segments, len(c))
	}
	buf := make([]byte, Channels)
	for i, col := range c {
		buf[i*3], buf[i*3+1], buf[i*3+2] = col.R, col.G, col.B
	}
	return r.SetFrame(buf)
}

func (r *Ring) SetSegment(n int, c Color) error {
	if n < 0 || n >= Segments {
		return fmt.Errorf("led: segment %d out of range 0-%d", n, Segments-1)
	}
	all := make([]Color, Segments)
	all[n] = c
	return r.SetSegments(all)
}

func (r *Ring) SetAll(c Color) error {
	all := make([]Color, Segments)
	for i := range all {
		all[i] = c
	}
	return r.SetSegments(all)
}

type Ring struct {
	Path string

	absent bool

	mu   sync.Mutex
	last []byte
}

func New() *Ring { return &Ring{Path: DefaultPath} }

func NewAbsent() *Ring { return &Ring{absent: true} }

func (r *Ring) Present() bool { return !r.absent }

func (r *Ring) path() string {
	if r.Path == "" {
		return DefaultPath
	}
	return r.Path
}

func (r *Ring) attr(name string) string { return r.path() + "/" + name }

func (r *Ring) SetFrame(vals []byte) error {
	if len(vals) != Channels {
		return fmt.Errorf("led: need %d channels, got %d", Channels, len(vals))
	}
	if r.absent {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if bytes.Equal(vals, r.last) {
		return nil
	}
	if err := os.WriteFile(
		r.attr("frame"),
		[]byte(hex.EncodeToString(vals)+"\n"),
		0o644,
	); err != nil {
		return err
	}

	r.last = append(r.last[:0], vals...)
	return nil
}

func (r *Ring) Forget() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.last = nil
}

func (r *Ring) Frame() ([]byte, error) {
	if r.absent {
		return make([]byte, Channels), nil
	}
	b, err := os.ReadFile(r.attr("frame"))
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(strings.TrimSpace(string(b)))
}

func (r *Ring) Fill(v byte) error {
	buf := make([]byte, Channels)
	for i := range buf {
		buf[i] = v
	}
	return r.SetFrame(buf)
}

func (r *Ring) Off() error { return r.Fill(0) }

func (r *Ring) Current() (int, error) {
	if r.absent {
		return 0, nil
	}
	b, err := os.ReadFile(r.attr("led_current"))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

func (r *Ring) SetCurrent(v int) error {
	if r.absent {
		return nil
	}
	return os.WriteFile(r.attr("led_current"), []byte(strconv.Itoa(v)), 0o644)
}

func (r *Ring) SetBootAnimation(on bool) error {
	if r.absent {
		return nil
	}
	v := "0"
	if on {
		v = "1"
	}
	return os.WriteFile(r.attr("boot_animation"), []byte(v), 0o644)
}
