package dns

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/ygelfand/echolocal/internal/android/prop"
)

// dhcpcd's hook sets dhcp.wlan0.dns* from the lease; no framework promotes them to net.dns*.
var props = []string{
	"net.dns1", "net.dns2", "net.dns3", "net.dns4",
	"dhcp.wlan0.dns1", "dhcp.wlan0.dns2", "dhcp.wlan0.dns3", "dhcp.wlan0.dns4",
}

const (
	timeout = 5 * time.Second

	fresh = time.Minute
)

func Use() {
	net.DefaultResolver = &net.Resolver{PreferGo: true, Dial: dial}
	slog.Info("resolver configured", "nameservers", nameservers())
}

func dial(ctx context.Context, network, _ string) (net.Conn, error) {
	servers := nameservers()
	if len(servers) == 0 {
		return nil, fmt.Errorf("dns: the device reports no nameservers")
	}

	var err error
	for _, server := range servers {
		d := net.Dialer{Timeout: timeout}

		var conn net.Conn
		if conn, err = d.DialContext(ctx, network, net.JoinHostPort(server, "53")); err == nil {
			return conn, nil
		}
	}
	return nil, err
}

var (
	mu     sync.Mutex
	cached []string
	read   time.Time
)

func nameservers() []string {
	mu.Lock()
	defer mu.Unlock()

	if cached != nil && time.Since(read) < fresh {
		return cached
	}

	var out []string
	for _, name := range props {
		v, err := prop.Get(name)
		if err != nil || v == "" || slices.Contains(out, v) {
			continue
		}
		out = append(out, v)
	}

	cached, read = out, time.Now()
	return out
}
