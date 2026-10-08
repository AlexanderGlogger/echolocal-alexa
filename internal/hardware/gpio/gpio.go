package gpio

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

const sysfsBase = "/sys/class/gpio"

// MediaTek pin indices; the Linux GPIO base is allocated down from ARCH_NR_GPIOS and can move.
const (
	// gpiolib never claims MTK pin 87; only a sysfs export reaches it.
	MutePin = 87

	// The mute line lights this LED; driven low is full brightness, high or input is dim.
	MuteLEDPin = 88
)

func chipBase() int {
	entries, err := os.ReadDir(sysfsBase)
	if err != nil {
		return 357
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "gpiochip") {
			continue
		}
		b, err := os.ReadFile(sysfsBase + "/" + e.Name() + "/base")
		if err != nil {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(strings.TrimSpace(string(b)), "%d", &n); err == nil {
			return n
		}
	}
	return 357
}

func Line(mtkPin int) int { return chipBase() + mtkPin }

type Pin struct{ N int }

func (p Pin) dir() string             { return fmt.Sprintf("%s/gpio%d", sysfsBase, p.N) }
func (p Pin) path(attr string) string { return p.dir() + "/" + attr }

func (p Pin) exported() bool {
	_, err := os.Stat(p.dir())
	return err == nil
}

func (p Pin) Export() error {
	if p.exported() {
		return nil
	}
	if err := os.WriteFile(sysfsBase+"/export", []byte(fmt.Sprint(p.N)), 0o644); err != nil {
		if !p.exported() {
			return err
		}
	}
	return nil
}

func (p Pin) SetDirection(d string) error {
	return os.WriteFile(p.path("direction"), []byte(d), 0o644)
}

func (p Pin) Direction() (string, error) {
	b, err := os.ReadFile(p.path("direction"))
	return strings.TrimSpace(string(b)), err
}

func (p Pin) Set(v bool) error {
	s := "0"
	if v {
		s = "1"
	}
	return os.WriteFile(p.path("value"), []byte(s), 0o644)
}

func (p Pin) Get() (bool, error) {
	b, err := os.ReadFile(p.path("value"))
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(b)) == "1", nil
}

type Mute struct{ pin Pin }

var (
	muteOnce sync.Once
	mute     *Mute
	muteErr  error

	ledOnce sync.Once
	muteLED *MuteLED
	ledErr  error
)

func Microphone() (*Mute, error) {
	muteOnce.Do(func() { mute, muteErr = NewMute() })
	return mute, muteErr
}

func LED() (*MuteLED, error) {
	ledOnce.Do(func() { muteLED, ledErr = NewMuteLED() })
	return muteLED, ledErr
}

func NewMute() (*Mute, error) {
	p := Pin{N: Line(MutePin)}
	if err := p.Export(); err != nil {
		return nil, fmt.Errorf("export mute line: %w", err)
	}
	// Exporting resets direction to "in", which floats the line.
	cur, err := p.Get()
	if err != nil {
		return nil, err
	}
	if err := p.SetDirection("out"); err != nil {
		return nil, fmt.Errorf("set mute line to output: %w", err)
	}
	if err := p.Set(cur); err != nil {
		return nil, err
	}
	return &Mute{pin: p}, nil
}

func (m *Mute) Set(muted bool) error { return m.pin.Set(muted) }

func (m *Mute) Get() (bool, error) { return m.pin.Get() }

func (m *Mute) Toggle() (bool, error) {
	cur, err := m.Get()
	if err != nil {
		return false, err
	}
	next := !cur
	return next, m.Set(next)
}

type MuteLED struct{ pin Pin }

func NewMuteLED() (*MuteLED, error) {
	p := Pin{N: Line(MuteLEDPin)}
	if err := p.Export(); err != nil {
		return nil, fmt.Errorf("export mute LED line: %w", err)
	}
	if err := p.SetDirection("out"); err != nil {
		return nil, fmt.Errorf("set mute LED line to output: %w", err)
	}
	return &MuteLED{pin: p}, nil
}

func (l *MuteLED) SetBright(bright bool) error { return l.pin.Set(!bright) }

func (l *MuteLED) Bright() (bool, error) {
	driven, err := l.pin.Get()
	if err != nil {
		return false, err
	}
	return !driven, nil
}
