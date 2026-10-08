package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/ygelfand/go-esphome-device/mdns"
	netaddress "github.com/ygelfand/libcountertop/pkg/network/address"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/layout"
)

const mdnsRetry = 3 * time.Second

// The mdns library falls back to loopback when nothing routable exists.
func (a *API) advertise(ctx context.Context, port int) {
	for attempt := 1; ; attempt++ {
		ips := netaddress.Addresses()
		if len(ips) == 0 {
			if attempt == 1 {
				slog.Info("waiting for an address before advertising over mdns")
			}
			if !pause(ctx, mdnsRetry) {
				return
			}
			continue
		}

		adv, err := mdns.Advertise(mdns.Config{
			Name:          a.name,
			FriendlyName:  a.srv.Info.FriendlyName,
			Port:          port,
			MACAddress:    a.srv.Info.MACAddress,
			Version:       a.srv.Info.Version,
			Platform:      layout.Platform,
			Board:         config.Get().Device.Board.Codename,
			Encrypted:     a.srv.PSK != nil && !a.srv.PSK.IsZero(),
			Provisionable: a.srv.PSK != nil && a.srv.PSK.IsZero(),
			IPs:           ips,
		})
		if err != nil {
			if attempt == 1 {
				slog.Warn("mdns advertise failed, retrying", "err", err)
			} else {
				slog.Debug("mdns advertise failed", "attempt", attempt, "err", err)
			}
			if !pause(ctx, mdnsRetry) {
				return
			}
			continue
		}

		slog.Info(
			"advertising over mdns",
			"name",
			a.name,
			"port",
			port,
			"addrs",
			netaddress.AddressKey(ips),
		)

		for netaddress.AddressKey(netaddress.Addresses()) == netaddress.AddressKey(ips) {
			if !pause(ctx, mdnsRetry) {
				adv.Close()
				return
			}
		}
		adv.Close()
		slog.Info("addresses changed, re-advertising over mdns", "was", netaddress.AddressKey(ips))
	}
}

func pause(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
