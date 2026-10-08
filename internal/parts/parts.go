package parts

import (
	"bytes"
	_ "embed"
	"log/slog"
	"os"

	"github.com/ygelfand/echolocal/internal/android/prop"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/update"
)

//go:embed echolocal.rc
var initRC []byte

type Part struct {
	Name    string
	Path    string
	Data    []byte
	Mode    os.FileMode
	Service string
	Boot    bool
}

var InitRC = Part{Name: "service", Path: layout.InitRC, Data: initRC, Mode: 0o644, Boot: true}

var registered []Part

func register(p Part) { registered = append(registered, p) } //nolint:unused

func All() []Part { return registered }

var (
	writable = update.Writable
	restart  = prop.Restart
	label    = setLabel
)

func Ensure() (rebootPending bool) { return ensure(registered) }

func ensure(ps []Part) (rebootPending bool) {
	var stale []Part
	for _, p := range ps {
		if have, err := os.ReadFile(p.Path); err == nil && bytes.Equal(have, p.Data) {
			continue
		}
		stale = append(stale, p)
	}
	if len(stale) == 0 {
		return false
	}

	if err := writable(true); err != nil {
		slog.Error("remounting to update parts failed", "err", err)
		return false
	}
	defer func() {
		if err := writable(false); err != nil {
			slog.Error("remounting read-only failed", "err", err)
		}
	}()

	for _, p := range stale {
		if err := replace(p); err != nil {
			slog.Error("updating a part failed", "part", p.Name, "path", p.Path, "err", err)
			continue
		}
		slog.Info("part updated", "part", p.Name, "path", p.Path)

		if p.Service != "" {
			if err := restart(p.Service); err != nil {
				slog.Error(
					"restarting a part failed",
					"part",
					p.Name,
					"service",
					p.Service,
					"err",
					err,
				)
			}
		}
		if p.Boot {
			slog.Warn("init reads this at the next boot", "part", p.Name)
			rebootPending = true
		}
	}
	return rebootPending
}

func replace(p Part) error {
	tmp := p.Path + ".new"
	if err := os.WriteFile(tmp, p.Data, p.Mode); err != nil {
		return err
	}
	if err := os.Chmod(tmp, p.Mode); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := label(tmp); err != nil {
		slog.Warn("labelling a part failed", "path", tmp, "err", err)
	}
	if err := os.Rename(tmp, p.Path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
