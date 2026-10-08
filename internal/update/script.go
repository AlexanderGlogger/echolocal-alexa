package update

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/layout"
)

// Amazon's animation scripts run a ledctrl that waits forever on a binder service echod lacks.

var Rollback = fmt.Sprintf(`#!/system/bin/sh
# Installed by EchoLocal, replacing a ledctrl call that waits forever on a binder service echod does
# not publish. echod drives the ring instead.
#
# It also puts back the binary an update replaced. echod renames %[1]s away once it has run long enough
# to be believed, so finding one here means a boot happened while an update was still on trial.
PREV=
if [ -f %[1]s ]; then
    PREV=%[1]s
elif [ -f %[5]s ]; then
    PREV=%[5]s
fi
if [ -n "$PREV" ]; then
    WAS=$(cat %[3]s 2>/dev/null)
    log -t echolocal "rolling back to the previous echod: an update did not settle (${WAS:-unknown})"
    mount -o remount,rw /system
    cat "$PREV" > %[2]s && rm -f "$PREV"
    mount -o remount,ro /system
    rm -f %[3]s
    setprop %[4]s "${WAS:-1}"
fi
exit 0
`, layout.PrevBinary, layout.Binary, layout.UpdatingPath, layout.RolledBackProp, layout.SystemPrevBinary)

var Stub = `#!/system/bin/sh
# Installed by EchoLocal, replacing a ledctrl call that waits forever on a binder service echod does
# not publish. echod drives the ring instead.
exit 0
`

// The other hook fires on amazon.headless.BOOT_COMPLETED, about a minute into the boot.
var hooks = func() []string { return board.Detect().BootHooks }

func Script(path string) string {
	if h := hooks(); len(h) > 0 && path == h[0] {
		return Rollback
	}
	return Stub
}

func Ensure() {
	var stale []string
	for _, path := range hooks() {
		want := Script(path)
		if have, err := os.ReadFile(path); err == nil && string(have) == want {
			continue
		}
		stale = append(stale, path)
	}
	if len(stale) == 0 {
		return
	}

	if err := writable(true); err != nil {
		slog.Error("remounting to update the boot hooks failed", "err", err)
		return
	}
	defer func() {
		if err := writable(false); err != nil {
			slog.Error("remounting read-only failed", "err", err)
		}
	}()

	for _, path := range stale {
		if err := write(path, Script(path)); err != nil {
			slog.Error("updating a boot hook failed", "path", path, "err", err)
			continue
		}
		slog.Info("boot hook updated", "path", path)
	}
}

func write(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		return err
	}
	if out, err := exec.Command("chcon", layout.OurLabel, path).CombinedOutput(); err != nil {
		slog.Warn("labelling a boot hook failed", "path", path, "err", err, "output", string(out))
	}
	return nil
}
