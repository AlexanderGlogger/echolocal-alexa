package privacy

import (
	"time"

	"github.com/ygelfand/echolocal/internal/hardware/gpio"
)

type line struct{ *gpio.Mute }

func (line) HardwareToggles() bool { return false }

func (line) Lag() time.Duration { return 0 }

func exported() (Mute, error) {
	m, err := gpio.Microphone()
	if err != nil {
		return nil, err
	}
	return line{m}, nil
}

func exportedLED() (LED, error) {
	l, err := gpio.LED()
	if err != nil {
		return nil, err
	}
	return l, nil
}
