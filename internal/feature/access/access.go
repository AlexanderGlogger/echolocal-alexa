package access

import (
	"log/slog"
	"sync"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/system/access"

	"github.com/ygelfand/echolocal/internal/android/firewall"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		func() componentAccess { return componentAccess{Get()} },
	)
}

// The boot image sets service.adb.tcp.port to this.
const adbPort = 5555

type componentAccess struct{ *access.Access }

// The firewall chain is empty after a reboot but not after an echod restart.
func (c componentAccess) Restore(cfg config.Config) { c.Access.Restore(cfg.Access) }

var get = sync.OnceValue(func() *access.Access {
	a := access.New(access.Options{Settings: config.AccessSection, ADB: reachable})
	component.Settings.Add(a.Controls())
	return a
})

func Get() *access.Access { return get() }

func reachable(on bool) error {
	change, what := func() error { return firewall.Close(firewall.ADB) }, "closing"
	if on {
		change, what = func() error { return firewall.Open(firewall.ADB, adbPort) }, "opening"
	}
	if err := change(); err != nil {
		slog.Error(what+" the adb port failed", "port", adbPort, "err", err)
		return err
	}
	slog.Warn("remote adb", "open", on, "port", adbPort)
	return nil
}
