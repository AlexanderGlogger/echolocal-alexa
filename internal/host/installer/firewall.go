package installer

import (
	"fmt"
	"strings"

	"github.com/ygelfand/echolocal/internal/host/device"
	"github.com/ygelfand/echolocal/internal/layout"
)

// Amazon's firewall.sh sets INPUT policy DROP and allows only its own ports.
var firewallHook = fmt.Sprintf(`#!/system/bin/sh
# Installed by EchoLocal. Amazon's firewall.sh runs this hook if it exists.
IPTABLES=/system/bin/iptables
$IPTABLES -A INPUT -i wlan0 -p tcp --dport %d -j ACCEPT
`, layout.Port)

func installFirewallHook(r *run) (string, bool, error) {
	existing, err := r.d.Exists(layout.FirewallHook)
	if err != nil {
		return "", false, err
	}
	if existing {
		current, err := r.d.ReadFile(layout.FirewallHook)
		if err != nil {
			return "", false, err
		}
		if string(current) == firewallHook {
			if err := applyFirewallRule(r.d); err != nil {
				return "", false, err
			}
			return "already installed", true, nil
		}
		if !strings.Contains(string(current), "EchoLocal") {
			return "", false, fmt.Errorf("%s exists and is not ours", layout.FirewallHook)
		}
	}

	if err := r.d.WriteFile(layout.FirewallHook, []byte(firewallHook), 0o755); err != nil {
		return "", false, err
	}
	if err := r.d.Chcon(layout.OurLabel, layout.FirewallHook); err != nil {
		return "", false, err
	}
	if err := applyFirewallRule(r.d); err != nil {
		return "", false, err
	}
	return fmt.Sprintf("tcp/%d open via %s", layout.Port, layout.FirewallHook), false, nil
}

func applyFirewallRule(d *device.Device) error {
	rule := fmt.Sprintf("INPUT -i wlan0 -p tcp --dport %d -j ACCEPT", layout.Port)
	if _, code, err := d.ShellCode("iptables -C " + rule); err != nil {
		return err
	} else if code == 0 {
		return nil
	}
	_, err := d.Shell("iptables -A " + rule)
	return err
}
