package timesync

import (
	"net"
	"sync"
	"time"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/service"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"
	"github.com/ygelfand/libcountertop/pkg/settings/timezone"
	"github.com/ygelfand/libcountertop/pkg/system/clock"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dhcp"
	"github.com/ygelfand/echolocal/internal/layout"
)

func init() {
	component.Register(sharedcomponent.Network, Get, sharedcomponent.Order(20),
		sharedcomponent.Supervise(service.Restart(5*time.Second, time.Minute)))
}

type Time struct{ *clock.Clock }

var (
	once   sync.Once
	shared *Time
)

func Get() *Time {
	once.Do(func() {
		shared = &Time{clock.New(clock.Options{
			Servers: servers,
			Built:   layout.BuildDate,
			RTC:     "/dev/rtc0",
			Zone: timezone.PolicyOptions{
				Read:   func() schema.Time { return config.Get().Time },
				Home:   func(s string) error { return config.Set().Time().Home(s) },
				Chosen: func(s string) error { return config.Set().Time().Chosen(s) },
			},
		})}
	})
	return shared
}

func (t *Time) Restore(config.Config) { t.Clock.Restore() }

func servers() []net.IP {
	if lease := dhcp.Get().Lease(); lease != nil {
		return lease.NTP
	}
	return nil
}
