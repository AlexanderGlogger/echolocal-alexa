//go:build board_checkers || board_cronos || board_rook

// Package a2dp is the device as a Bluetooth speaker.
//
// The pieces all existed and none of them were joined: internal/lib/bt decides what a phone's bytes
// mean, internal/lib/bt/pair decides a pairing, internal/lib/bt/sbc turns frames into samples, and
// internal/hardware/ble carries the line. This is the wiring, and nothing else — no protocol lives
// here, so a fault is in one of the parts rather than spread between them.
package a2dp

import (
	"sync"
	"time"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/volume"
	"github.com/ygelfand/echolocal/internal/hardware/ble"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/service"
	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	sharedspeaker "github.com/ygelfand/libcountertop/pkg/bluetooth/a2dp"
)

var once sync.Once
var shared *sharedspeaker.Sink

func build() *sharedspeaker.Sink {
	return sharedspeaker.New(sharedspeaker.Dependencies{
		Queue:            speaker.Get(),
		Radio:            ble.Get,
		Bonds:            layout.StateDir + "/bonds",
		CaptureDirectory: "/data/local/tmp",
		Name:             func() string { return config.Get().Device.Name },
		Settings:         config.BluetoothSection,
		Volume:           sharedvolume.For(volume.Get(), config.StreamMedia),
		Media:            media.Get(),
		Arbitration:      speaker.Sound().Backgrounds(),
		Ducking:          config.MediaSection,
	})
}
func Get() *sharedspeaker.Sink { once.Do(func() { shared = build() }); return shared }
func init() {
	component.Register(component.Device, Get, component.Order(70), component.Needs(board.Panel),
		component.Supervise(service.Restart(5*time.Second, time.Minute)))
}
