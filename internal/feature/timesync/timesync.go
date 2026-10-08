package timesync

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	esphome "github.com/ygelfand/go-esphome-device"
	"github.com/ygelfand/go-esphome-device/api"
	"google.golang.org/protobuf/proto"

	"github.com/ygelfand/libcountertop/pkg/hook"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/settings/schema"
	zonepolicy "github.com/ygelfand/libcountertop/pkg/settings/timezone"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/layout"
)

// The protocol has the device ask for the time and the client answer.

func init() {
	component.Register(sharedcomponent.Device, Get, sharedcomponent.Order(1))
}

const step = time.Second

type Sync struct {
	Before, After time.Time
}

func (s Sync) Stepped() bool { return !s.After.Equal(s.Before) }

type Time struct {
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

func (t *Time) Handle(_ context.Context, conn *esphome.Conn, msg proto.Message) error {
	if reply, ok := msg.(*api.GetTimeResponse); ok {
		t.fromHome(reply.GetEpochSeconds())
		t.zoneFromHome(reply.GetTimezone())
		return nil
	}

	if t.shouldAsk(conn) {
		if err := conn.Send(&api.GetTimeRequest{}); err != nil {
			slog.Debug("asking home assistant the time failed", "err", err)
		}
	}
	return nil
}

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
		t.zonePolicy = zonepolicy.NewPolicy(
			zonepolicy.PolicyOptions{
				Read:   func() schema.Time { return config.Get().Time },
				Home:   func(s string) error { return config.Set().Time().Home(s) },
				Chosen: func(s string) error { return config.Set().Time().Chosen(s) },
			},
		)
	})
	return t.zonePolicy
}

func (t *Time) Valid() bool {
	if t.set.Load() {
		return true
	}
	built, err := time.Parse(time.RFC3339, layout.BuildDate)
	return err != nil || time.Now().After(built)
}

func (c *Time) WatchSteps(changed func(time.Duration)) func() {
	return c.Synced.Listen(func(s Sync) {
		if s.Stepped() {
			changed(s.After.Sub(s.Before))
		}
	})
}
