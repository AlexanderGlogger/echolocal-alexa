package services

import "strings"

var Disabled = []string{
	"mixer",
	"BTSinkPlayer",
	"btmanagerd",
	"blemesh_service",

	// puffin is the Alexa client.
	"puffin",
	"puffinmrmd",
	"ahe",
	"smarthomed",
	"commsd",
	"shs",
	"amakit_server",
	"tokend",
	"provisionerd",
	"oobed_on_boot",
	"UdssCampSvc",
	"assetmgrd",
	"dacd",
	"uxeventd",

	"adepd",

	// fosflags strips adb out of persist.sys.usb.config on every boot.
	"fosflags",

	"ace_metricd",
	"aceusagestatd",
	"usagestat_protod",
	"acedropboxd",
	"minerva_service",
	"trackerd",
	"logmgr",
	"ace_coex_metric",
	"otad",
	"ace_otad",
	"halo-bssid-scan",

	"acebuttond",
	"aceinputmanager",
	"ace_sensorsd",

	"perfmonitord",
	"perfrecoveryd",

	"factory-reset",
	"avahi-daemon",
	"neo-coordinator",
	"neo-init",
	"fireos-dha",
}

// dhcpcd ships disabled and no rc starts it.
var Enabled = []string{"dhcpcd-wlan0"}

func Enable(rc string, names map[string]bool) (string, []string) {
	lines := strings.Split(rc, "\n")

	var (
		out     = make([]string, 0, len(lines))
		changed []string
		inside  string
	)

	for _, line := range lines {
		if name, ok := serviceName(line); ok {
			inside = ""
			if names[name] {
				inside = name
			}
			out = append(out, line)
			continue
		}

		if inside != "" {
			switch trimmed := strings.TrimSpace(line); {
			case trimmed == "":
			case trimmed == "disabled":
				changed = append(changed, inside)
				continue
			case !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t"):
				inside = ""
			}
		}
		out = append(out, line)
	}

	return strings.Join(out, "\n"), changed
}

func DisabledSet() map[string]bool { return Set(Disabled) }
func EnabledSet() map[string]bool  { return Set(Enabled) }

func Set(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		out[name] = true
	}
	return out
}

// disabled stops a class start, not an explicit "start name" in an on-block.
func Disable(rc string, names map[string]bool) (string, []string) {
	lines := strings.Split(rc, "\n")

	var (
		out     = make([]string, 0, len(lines)+len(names))
		changed []string
	)

	for i := 0; i < len(lines); i++ {
		if name, ok := started(lines[i]); ok && names[name] {
			out = append(out, "#"+lines[i])
			changed = append(changed, "start "+name)
			continue
		}

		out = append(out, lines[i])

		name, ok := serviceName(lines[i])
		if !ok || !names[name] {
			continue
		}

		for strings.HasSuffix(strings.TrimRight(lines[i], " \t"), `\`) && i+1 < len(lines) {
			i++
			out = append(out, lines[i])
		}

		if blockHas(lines[i+1:], "disabled") {
			continue
		}
		out = append(out, "    disabled")
		changed = append(changed, name)
	}

	return strings.Join(out, "\n"), changed
}

// init refuses a service with no domain transition: "does not have a SELinux domain defined".
const Domain = "u:r:su:s0"

func AsRoot(rc, name string) (string, []string) {
	lines := strings.Split(rc, "\n")

	var (
		out     = make([]string, 0, len(lines)+1)
		changed []string
		inside  bool
	)

	for i := 0; i < len(lines); i++ {
		if at, ok := serviceName(lines[i]); ok {
			out = append(out, lines[i])

			if inside = at == name; !inside {
				continue
			}
			if !blockHas(lines[i+1:], "seclabel") {
				out = append(out, "    seclabel "+Domain)
				changed = append(changed, "seclabel "+Domain)
			}
			continue
		}

		if inside {
			switch fields := strings.Fields(lines[i]); {
			case len(fields) == 0:
			case fields[0] == "user", fields[0] == "group":
				changed = append(changed, strings.TrimSpace(lines[i]))
				continue
			case !strings.HasPrefix(lines[i], " ") && !strings.HasPrefix(lines[i], "\t"):
				inside = false
			}
		}
		out = append(out, lines[i])
	}

	return strings.Join(out, "\n"), changed
}

func started(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) != 2 || fields[0] != "start" {
		return "", false
	}
	return fields[1], true
}

func serviceName(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 || fields[0] != "service" {
		return "", false
	}
	return fields[1], true
}

func blockHas(rest []string, keyword string) bool {
	for _, line := range rest {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
		case !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t"):
			return false
		case trimmed == keyword, strings.HasPrefix(trimmed, keyword+" "):
			return true
		}
	}
	return false
}

func Gated(rc, name string) (string, []string) {
	lines := strings.Split(rc, "\n")

	out := make([]string, 0, len(lines)+1)
	var changed []string

	for i, line := range lines {
		out = append(out, line)

		at, ok := serviceName(line)
		if !ok || at != name || blockHas(lines[i+1:], "disabled") {
			continue
		}
		out = append(out, "    disabled")
		changed = append(changed, name+" disabled")
	}
	return strings.Join(out, "\n"), changed
}
