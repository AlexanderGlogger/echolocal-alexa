// Package timesync takes the time and the time zone from Home Assistant.
package timesync

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/layout"
	esphome "github.com/ygelfand/go-esphome-device"
	"github.com/ygelfand/go-esphome-device/api"
	"github.com/ygelfand/libcountertop/pkg/hook"
	zonepolicy "github.com/ygelfand/libcountertop/pkg/settings/timezone"
	"google.golang.org/protobuf/proto"
)

// Home Assistant knows the time and is already talking to this device, so it is the source: it
// works on a network with no route out, and it is what an ESPHome device with
// `time: platform: homeassistant` uses.
//
// The protocol has the device ask and the client answer, so this both sends the request and takes
// the reply.

func init() {
	component.Register(component.Device, Get, component.Order(1))
}

// step is the drift worth stepping the clock for.
const step = time.Second

// Sync is the clock before and after one answer from Home Assistant.
type Sync struct {
	Before, After time.Time
}

// Stepped reports whether the clock was moved.
func (s Sync) Stepped() bool { return !s.After.Equal(s.Before) }

type Time struct {
	// Synced carries every answer from Home Assistant.
	Synced hook.Hook[Sync]

	mu         sync.Mutex
	asked      *esphome.Conn
	zoneOnce   sync.Once
	zonePolicy *zonepolicy.Policy

	set atomic.Bool
}

var (
	once   sync.Once
	shared *Time
)

func Get() *Time {
	once.Do(func() { shared = &Time{} })
	return shared
}

func (t *Time) Name() string { return "time" }

// Handle takes the answer, and asks the question the first time a client says anything.
func (t *Time) Handle(_ context.Context, conn *esphome.Conn, msg proto.Message) error {
	if reply, ok := msg.(*api.GetTimeResponse); ok {
		t.fromHome(reply.GetEpochSeconds())
		t.zoneFromHome(reply.GetTimezone())
		return nil
	}

	// Any message means a client is there to ask. Once per connection: one that reconnects is
	// worth asking again, and after a long disconnection it is the better source.
	if t.shouldAsk(conn) {
		if err := conn.Send(&api.GetTimeRequest{}); err != nil {
			slog.Debug("asking home assistant the time failed", "err", err)
		}
	}
	return nil
}

// shouldAsk reports whether this connection has not been asked yet, and records that it is about to
// be.
func (t *Time) shouldAsk(conn *esphome.Conn) bool {
	if conn == nil {
		return false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if conn == t.asked {
		return false
	}
	t.asked = conn
	return true
}

// fromHome takes an epoch Home Assistant reported.
func (t *Time) fromHome(epoch uint32) {
	if epoch == 0 {
		return
	}

	at := time.Unix(int64(epoch), 0)
	before := time.Now()
	if offset := at.Sub(before); offset < step && offset > -step {
		t.set.Store(true)
		t.Synced.Emit(Sync{Before: before, After: before})
		return
	}

	if err := setClock(at); err != nil {
		slog.Warn("could not take the time from home assistant", "err", err)
		return
	}
	t.set.Store(true)
	slog.Info("clock set", "source", "home assistant", "now", at, "was", before.Round(time.Second))
	t.Synced.Emit(Sync{Before: before, After: at})
}

// zoneFromHome takes the zone Home Assistant reported, which is what turns a clock face from UTC
// into the time on the wall.
func (t *Time) zoneFromHome(said string) { t.policy().FromHome(said) }
func (t *Time) Zone() string {
	c := config.Get().Time
	if c.Chosen != "" {
		return c.Chosen
	}
	return c.Home
}
func (t *Time) ChosenZone() string         { return t.policy().Zone() }
func (t *Time) SetZone(s string)           { t.policy().SetZone(s) }
func (t *Time) Restore(config.Config)      { t.policy().Restore() }
func (t *Time) Entities() []esphome.Entity { return []esphome.Entity{t.policy().Select} }
func (t *Time) policy() *zonepolicy.Policy {
	t.zoneOnce.Do(func() {
		t.zonePolicy = zonepolicy.NewPolicy(zonepolicy.PolicyOptions{Read: func() config.Time { return config.Get().Time }, Home: func(s string) error { return config.Set().Time().Home(s) }, Chosen: func(s string) error { return config.Set().Time().Chosen(s) }})
	})
	return t.zonePolicy
}

// Valid reports whether the time is worth showing: Home Assistant has set it, or it is already later
// than this build.
func (t *Time) Valid() bool {
	if t.set.Load() {
		return true
	}
	built, err := time.Parse(time.RFC3339, layout.BuildDate)
	return err != nil || time.Now().After(built)
}
