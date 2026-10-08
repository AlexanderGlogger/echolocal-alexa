package wifi

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ygelfand/echolocal/internal/layout"
)

const Interface = layout.WifiIface

func Loaded() bool {
	_, err := os.Stat("/sys/class/net/" + Interface)
	return err == nil
}

// init insmods the driver from the boot image, possibly after echod starts.
const appear = 15 * time.Second

func Load() error {
	deadline := time.Now().Add(appear)
	for {
		if Loaded() {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s did not appear within %s", Interface, appear)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func MAC() (string, error) {
	b, err := os.ReadFile("/sys/class/net/" + Interface + "/address")
	if err != nil {
		return "", fmt.Errorf("reading the wlan address: %w", err)
	}
	mac := strings.TrimSpace(string(b))
	if mac == "" || mac == "00:00:00:00:00:00" {
		return "", fmt.Errorf("the wlan address reads %q", mac)
	}
	return mac, nil
}
