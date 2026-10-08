package installer

import (
	"strings"

	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/update"
)

// ledctrl blocks in getService for the LedController binder service echod does not publish.

func disableBootAnimation(r *run) (string, bool, error) {
	hooks := r.board().BootHooks
	if len(hooks) == 0 {
		return "no boot hooks on this board", true, nil
	}

	var done []string
	for _, path := range hooks {
		replaced, err := stubScript(r, path)
		if err != nil {
			return "", false, err
		}
		if replaced {
			done = append(done, path)
		}
	}
	if len(done) == 0 {
		return "already stubbed", true, nil
	}

	r.reboot = true
	return strings.Join(done, ", "), false, nil
}

func stubScript(r *run, path string) (bool, error) {
	want := update.Script(path)

	current, err := r.d.ReadFile(path)
	if err == nil && string(current) == want {
		return false, nil
	}

	if err := r.d.WriteFile(path, []byte(want), 0o755); err != nil {
		return false, err
	}
	return true, r.d.Chcon(layout.OurLabel, path)
}
