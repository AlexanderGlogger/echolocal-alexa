package installer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ygelfand/echolocal/internal/host/device"
	"github.com/ygelfand/echolocal/internal/layout"
)

var ErrNoName = errors.New("no device name")

func installName(r *run) (string, bool, error) {
	existing, err := ReadName(r.d)
	if err != nil {
		return "", false, err
	}
	if r.cfg.Name == "" {
		if existing != "" {
			return existing, true, nil
		}
		return "", false, fmt.Errorf("%w: pass --name", ErrNoName)
	}
	if existing == r.cfg.Name {
		return existing, true, nil
	}
	if err := validName(r.cfg.Name); err != nil {
		return "", false, err
	}

	if _, err := r.d.Shell("mkdir -p " + layout.StateDir); err != nil {
		return "", false, err
	}
	if err := r.d.WriteFile(layout.NamePath, []byte(r.cfg.Name+"\n"), 0o644); err != nil {
		return "", false, err
	}
	return r.cfg.Name, false, nil
}

func ReadName(d *device.Device) (string, error) {
	have, err := d.Exists(layout.NamePath)
	if err != nil || !have {
		return "", err
	}
	b, err := d.ReadFile(layout.NamePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func SuggestName(d *device.Device) string {
	on := BoardOf(d)

	out, err := d.Shell("cat " + layout.MACPath)
	if err != nil {
		return on.DefaultName
	}
	return on.NameFromMAC(out)
}

func validName(name string) error {
	slug := layout.Slug(name)
	if slug == "" {
		return fmt.Errorf("name %q has no letters or digits to build a hostname from", name)
	}
	if len(slug) > layout.MaxNodeName {
		return fmt.Errorf("name %q becomes %q, %d characters, and the limit is %d",
			name, slug, len(slug), layout.MaxNodeName)
	}
	return nil
}
