package wifi

import (
	"fmt"
	"strings"
	"time"

	"github.com/ygelfand/echolocal/internal/host/device"
	"github.com/ygelfand/echolocal/internal/layout"
)

const (
	Service = "wpa_supplicant"

	// update_config=1 lets a join persist.
	config = "ctrl_interface=" + layout.WifiSockets + "\nupdate_config=1\n"
)

func Up(d *device.Device) (string, error) {
	var did []string

	written, err := writeConfig(d)
	if err != nil {
		return "", err
	}
	if written {
		did = append(did, "configuration written")
	}

	started, err := startSupplicant(d)
	if err != nil {
		return "", err
	}
	if started {
		did = append(did, "supplicant started")
	}

	if len(did) == 0 {
		return "already up", nil
	}
	return strings.Join(did, ", "), nil
}

func writeConfig(d *device.Device) (bool, error) {
	has, err := d.Exists(layout.WifiConf)
	if err != nil || has {
		return false, err
	}

	if _, err := d.Shell("mkdir -p " + layout.WifiSockets); err != nil {
		return false, fmt.Errorf("wifi: %w", err)
	}
	if err := d.WriteFile(layout.WifiConf, []byte(config), 0o660); err != nil {
		return false, fmt.Errorf("wifi: writing %s: %w", layout.WifiConf, err)
	}

	for _, path := range []string{layout.WifiSockets, layout.WifiConf} {
		if _, err := d.Shell(
			fmt.Sprintf("chown %d:%d %s", layout.WifiUser, layout.WifiUser, path),
		); err != nil {
			return false, fmt.Errorf("wifi: %w", err)
		}
	}
	return true, nil
}

func startSupplicant(d *device.Device) (bool, error) {
	if answering(d) {
		return false, nil
	}

	if err := d.Setprop("ctl.start", Service); err != nil {
		return false, err
	}

	for range 20 {
		if answering(d) {
			return true, nil
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false, fmt.Errorf("wifi: the supplicant did not answer on %s", layout.WifiSockets)
}

func answering(d *device.Device) bool {
	out, err := shell{d}.Cmd("ping")
	return err == nil && strings.Contains(out, "PONG")
}
