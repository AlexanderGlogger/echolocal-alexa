package metrics

import (
	"path/filepath"
	"strconv"
)

// tsl2584tsv (IIO) or tsl2540 (vendor driver); reading either one's calibrated value crashes the device.
func (r Reader) LuxPath() string {
	for i := range 8 {
		at := r.path("sys/bus/iio/devices/iio:device"+strconv.Itoa(i)) + "/illuminance0_input"

		if _, err := number(at); err == nil {
			return at
		}
	}

	vendor, _ := filepath.Glob(r.path("sys/bus/i2c/devices/*/als_lux"))
	for _, at := range vendor {
		if _, err := number(at); err == nil {
			return at
		}
	}
	return ""
}

var Ambient func() Reading

func (r Reader) Lux(path string) Reading {
	if path == "" {
		if Ambient == nil {
			return Reading{}
		}
		return Ambient()
	}

	lux, err := number(path)
	if err != nil {
		return Reading{}
	}
	return known(lux)
}
