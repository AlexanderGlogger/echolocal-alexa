package privacy

import (
	"os"
	"strings"
	"time"
)

// Reading power_button_state in this directory reboots the device.
const (
	dir = "/sys/devices/soc/10010000.keypad/amz_privacy"

	state      = dir + "/privacy_state"
	cut        = dir + "/enable"
	release    = dir + "/shutdown_dialog_state"
	brightness = dir + "/privacy_brightness"
)

// Cutting and releasing are separate files and neither reports the result; privacy_state does.
type driver struct{}

func (driver) Get() (bool, error) { return reads(state, "1") }

// The driver's own handler runs before the key event is emitted.
func (driver) HardwareToggles() bool { return true }

// A cut is reported about 300 ms after the press, a release at once, on 6.5.7.4.
func (driver) Lag() time.Duration { return 500 * time.Millisecond }

// Releasing stands the privacy handler down; the button is dead until the second write.
func (d driver) Set(muted bool) error {
	switch is, err := d.Get(); {
	case err != nil:
		return err
	case is == muted:
		return nil
	case muted:
		return write(cut, "1")
	}

	if err := write(release, "1"); err != nil {
		return err
	}
	return write(release, "0")
}

func (d driver) Toggle() (bool, error) {
	is, err := d.Get()
	if err != nil {
		return false, err
	}
	return !is, d.Set(!is)
}

// The value that lights it fully is 0.
type driverLED struct{}

func (driverLED) SetBright(bright bool) error {
	if bright {
		return write(brightness, "0")
	}
	return write(brightness, "1")
}

func (driverLED) Bright() (bool, error) { return reads(brightness, "0") }

func write(path, value string) error { return os.WriteFile(path, []byte(value), 0o644) }

func reads(path, value string) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(b)) == value, nil
}
