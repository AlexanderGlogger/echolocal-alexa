package recording

import (
	"sync"

	"github.com/ygelfand/libcountertop/pkg/assistant/recordings"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/layout"
)

func init() {
	component.Register(sharedcomponent.Device, Get, sharedcomponent.Order(20))
}

type Store struct{ *recordings.Store }

var (
	once   sync.Once
	shared *Store
)

func Get() *Store {
	once.Do(func() {
		shared = &Store{recordings.New(recordings.Options{
			Dir:      layout.RecordingDir,
			Rate:     mic.Rate,
			Slots:    2,
			DeviceID: component.AssistantDevice,
			Keeps:    func(slot int) int { return config.Get().Wake.Slot(slot).Recordings },
			Save:     func(slot, count int) error { return config.Set().Wake(slot).Recordings(count) },
		})}
	})
	return shared
}

func (s *Store) Restore(config.Config) { s.Store.Restore() }
