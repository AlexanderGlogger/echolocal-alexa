package update

import (
	"log/slog"
	"os"
	"strings"

	"github.com/ygelfand/echolocal/internal/android/prop"
	"github.com/ygelfand/echolocal/internal/layout"
)

var (
	prev = layout.PrevBinary
	old  = layout.OldBinary

	binary     = layout.Binary
	aside      = layout.AsideBinary
	systemOld  = layout.SystemOldBinary
	systemPrev = layout.SystemPrevBinary

	mount = mountHolding(layout.Binary)

	writable = remount
)

func Writable(rw bool) error { return writable(rw) }

// Fire OS 5 mounts the system partition at /system; Fire OS 6 runs it as the root filesystem.
func mountHolding(path string) string {
	mounts, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return "/"
	}
	return mountIn(string(mounts), path)
}

func mountIn(mounts, path string) string {
	at := "/"
	for line := range strings.SplitSeq(mounts, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// /sys is a string prefix of /system/app.
		point := fields[1]
		if point == path || strings.HasPrefix(path, strings.TrimSuffix(point, "/")+"/") {
			if len(point) > len(at) {
				at = point
			}
		}
	}
	return at
}

var wanted = make(chan string, 1)

// Exiting with the amplifier still driven makes the speaker pop.
func Restart(why string) {
	select {
	case wanted <- why:
	default:
	}
}

func Wanted() <-chan string { return wanted }

// The kernel names the root source /dev/root, with no node, and adb remount only goes read-write.
func Remount(rw bool) error { return writable(rw) }

func OnTrial() bool { return exists(prev) || exists(systemPrev) }

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func Start() (onTrial, rebooting bool) {
	if !OnTrial() {
		return false, false
	}

	if was, _ := prop.Get(layout.TrialProp); was != "" {
		slog.Error(
			"this binary was already tried this boot and never settled, rebooting to go back",
			"trial",
			was,
			"prev",
			prev,
		)
		reboot()
		return true, true
	}

	if err := prop.Set(layout.TrialProp, "1"); err != nil {
		slog.Error("marking the update as on trial failed", "err", err)
	}
	slog.Warn("running an update on trial", "prev", prev)
	return true, false
}

func Commit() {
	if !OnTrial() {
		return
	}

	if err := keep(); err != nil {
		slog.Error("keeping an update failed", "err", err)
		return
	}
	if err := prop.Set(layout.TrialProp, ""); err != nil {
		slog.Error("clearing the trial property failed", "err", err)
	}
	slog.Info("update kept", "previous", old)
	tidy()
}

func keep() error {
	if exists(prev) {
		return os.Rename(prev, old)
	}
	if err := copyTo(systemPrev, old); err != nil {
		return err
	}
	return onSystem(func() error { return os.Remove(systemPrev) })
}

func tidy() {
	var stale []string
	for _, p := range []string{systemOld, aside} {
		if exists(p) {
			stale = append(stale, p)
		}
	}
	if len(stale) == 0 {
		return
	}
	err := onSystem(func() error {
		for _, p := range stale {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		slog.Error("clearing old binaries off /system failed", "err", err)
		return
	}
	slog.Info("cleared old binaries off /system", "removed", stale)
}

func onSystem(do func() error) error {
	if err := writable(true); err != nil {
		return err
	}
	err := do()
	if back := writable(false); back != nil {
		slog.Error("remounting read-only failed", "err", back)
	}
	return err
}

func RolledBack() string {
	was, err := prop.Get(layout.RolledBackProp)
	if err != nil || was == "" {
		return ""
	}

	slog.Warn("an update was rolled back before this boot", "version", was)
	if err := prop.Set(layout.RolledBackProp, ""); err != nil {
		slog.Error("clearing the rollback property failed", "err", err)
	}
	return was
}

var reboot = func() {
	if err := prop.Set("sys.powerctl", "reboot"); err != nil {
		slog.Error("asking init to reboot failed", "err", err)
	}
}

func Reboot(why string) {
	slog.Info("rebooting", "why", why)
	reboot()
}
