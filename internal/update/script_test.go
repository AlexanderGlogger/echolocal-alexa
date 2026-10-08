package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ygelfand/echolocal/internal/layout"
)

func TestOnlyTheEarlyHookRollsBack(t *testing.T) {
	early := Script(layout.StartAnimation)
	late := Script(layout.StopAnimation)

	if !strings.Contains(early, layout.PrevBinary) {
		t.Errorf("the boot hook does not mention %s:\n%s", layout.PrevBinary, early)
	}
	if strings.Contains(late, layout.PrevBinary) {
		t.Errorf("the hook that runs after the boot would roll back:\n%s", late)
	}
	if early == late {
		t.Error("both hooks have the same contents")
	}
}

func TestBothHooksExitCleanlyAndNeverCallLedctrl(t *testing.T) {
	for _, path := range hooks() {
		script := Script(path)

		if !strings.HasPrefix(script, "#!/system/bin/sh\n") {
			t.Errorf("%s has no shebang:\n%s", path, script)
		}
		if !strings.HasSuffix(script, "exit 0\n") {
			t.Errorf("%s does not end by exiting cleanly:\n%s", path, script)
		}
		if strings.Contains(script, "/system/bin/ledctrl") {
			t.Errorf("%s still calls ledctrl", path)
		}
	}
}

func TestRollbackRestoresAndReports(t *testing.T) {
	script := Script(layout.StartAnimation)

	for _, want := range []string{
		"mount -o remount,rw /system",
		"PREV=" + layout.PrevBinary,
		"PREV=" + layout.SystemPrevBinary,
		`cat "$PREV" > ` + layout.Binary,
		"mount -o remount,ro /system",
		"setprop " + layout.RolledBackProp,
	} {
		if !strings.Contains(script, want) {
			t.Errorf("the rollback is missing %q:\n%s", want, script)
		}
	}
}

func TestEnsureLeavesCurrentHooksAlone(t *testing.T) {
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "start.sh"), filepath.Join(dir, "stop.sh")}

	restore := hooks
	t.Cleanup(func() { hooks = restore; writable = remount })
	hooks = func() []string { return paths }

	var remounts int
	writable = func(bool) error { remounts++; return nil }

	for _, p := range paths {
		if err := os.WriteFile(p, []byte(Script(p)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	Ensure()
	if remounts != 0 {
		t.Errorf("remounted %d times with both hooks already current", remounts)
	}

	if err := os.WriteFile(paths[0], []byte("#!/system/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	Ensure()
	if remounts == 0 {
		t.Error("a stale hook was left alone")
	}
	if got, err := os.ReadFile(paths[0]); err != nil || string(got) != Script(paths[0]) {
		t.Errorf("the stale hook was not rewritten: %v", err)
	}
}
