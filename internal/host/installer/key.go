package installer

import (
	"errors"
	"fmt"
	"strings"

	esphome "github.com/ygelfand/go-esphome-device"
	"github.com/ygelfand/libcountertop/pkg/host/adb"

	"github.com/ygelfand/echolocal/internal/layout"
)

var ErrNoKey = errors.New("device has no encryption key")

func installKey(r *run) (string, bool, error) {
	existing, err := ReadKey(r.d)
	if err != nil {
		return "", false, err
	}
	if existing != "" {
		return "already provisioned", true, nil
	}
	if r.cfg.ZeroPSK {
		return "unprovisioned, waiting for Home Assistant to push a key", false, nil
	}

	k, err := esphome.GeneratePSK()
	if err != nil {
		return "", false, err
	}
	if err := writeKey(r.d, k.String()); err != nil {
		return "", false, err
	}
	return "generated", false, nil
}

func ReadKey(d *adb.Device) (string, error) {
	have, err := d.Exists(layout.KeyPath)
	if err != nil || !have {
		return "", err
	}
	b, err := d.ReadFile(layout.KeyPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func RotateKey(d *adb.Device) (string, error) {
	k, err := esphome.GeneratePSK()
	if err != nil {
		return "", err
	}
	if err := writeKey(d, k.String()); err != nil {
		return "", err
	}
	if err := d.Setprop("ctl.restart", BoardOf(d).ServiceName); err != nil {
		return "", err
	}
	return k.String(), nil
}

func writeKey(d *adb.Device, key string) error {
	if _, err := d.Shell("mkdir -p " + layout.StateDir); err != nil {
		return err
	}
	return d.WriteFile(layout.KeyPath, []byte(key+"\n"), 0o600)
}

func KeyOrError(d *adb.Device) (string, error) {
	k, err := ReadKey(d)
	if err != nil {
		return "", err
	}
	if k == "" {
		return "", fmt.Errorf("%w (%s missing)", ErrNoKey, layout.KeyPath)
	}
	return k, nil
}
