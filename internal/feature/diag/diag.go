package diag

import (
	"log/slog"
	"sync"
	"time"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/libcountertop/pkg/audio/upsample"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/system/diag"
	"github.com/ygelfand/libcountertop/pkg/system/metrics"

	"github.com/ygelfand/echolocal/internal/android/setup"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/bluetooth"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/hardware/mic"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/lib/wake"
)

func init() {
	component.Register(sharedcomponent.Network, Get, sharedcomponent.Order(90))
}

var Ambient func() metrics.Reading

// mtktscpu is the CPU, mtktswmt the wifi and Bluetooth combo chip.
var zones = []diag.Zone{
	{ID: "cpu_temperature", Name: "CPU temperature", Zone: "mtktscpu"},
	{ID: "radio_temperature", Name: "Radio temperature", Zone: "mtktswmt"},
}

type Diag struct {
	*diag.Diag

	testPlayback *esphome.Button
	lux          *esphome.Sensor
	roomLevel    *esphome.Sensor
	roomFloor    *esphome.Sensor
	ads          *esphome.Sensor
	color        *esphome.TextSensor
	boardName    *esphome.TextSensor

	luxPath string

	last struct {
		at       time.Time
		reports  uint64
		recorded bool
	}
}

var (
	once   sync.Once
	shared *Diag
)

func Get() *Diag {
	once.Do(func() {
		shared = &Diag{}
		shared.Diag = diag.New(diag.Options{
			Settings: config.DiagSection,
			StateDir: layout.StateDir,
			Version:  layout.VersionString(),
			Zones:    zones,
			Cache:    wake.Lib().CacheEntities(inUse),
			Cores:    setup.Present(),
			Hold:     setup.MinCores,
			Sample:   shared.sample,
		})
		component.Settings.Add(shared.Controls())
		shared.hardware()
		shared.playback()
		shared.identity()

		component.Subscribed.Listen(func(struct{}) { shared.Soon() })
	})
	return shared
}

func (d *Diag) Entities() []esphome.Entity {
	return append(d.Diag.Entities(),
		d.lux, d.roomLevel, d.roomFloor, d.ads, d.color, d.boardName, d.testPlayback)
}

func (d *Diag) Restore(c config.Config) { d.Diag.Restore(c.Diag) }

func (d *Diag) playback() {
	spk := speaker.Get()

	d.testPlayback = &esphome.Button{
		Base: esphome.Base{
			ObjectID: "test_playback",
			Name:     "Test playback",
			Icon:     "mdi:waveform",
			Category: esphome.CategoryDiagnostic,
		},
		OnPress: func() {
			slog.Info("test playback", "resampling", spk.Resampling(), "step", spk.Step())
			spk.PlayVoice(upsample.Sweep())
		},
	}
}

// The shell colour comes from the factory idme partition.
func (d *Diag) identity() {
	b := config.Get().Device.Board

	d.color = &esphome.TextSensor{
		Base: esphome.Base{
			ObjectID: "hardware_color",
			Name:     "Hardware color",
			Icon:     "mdi:palette",
			Category: esphome.CategoryDiagnostic,
		},
	}
	d.color.Set(b.Color(layout.Idme))

	d.boardName = &esphome.TextSensor{
		Base: esphome.Base{
			ObjectID: "hardware_board",
			Name:     "Hardware board",
			Icon:     "mdi:chip",
			Category: esphome.CategoryDiagnostic,
		},
	}
	d.boardName.Set(b.Codename)
}

func (d *Diag) hardware() {
	room := func(id, name, icon string) *esphome.Sensor {
		return &esphome.Sensor{
			Base: esphome.Base{
				ObjectID: id, Name: name, Icon: icon,
				DeviceID: component.DeviceMicrophone,
				Category: esphome.CategoryDiagnostic,
			},
			StateClass: esphome.StateClassMeasurement,
			Unit:       "dBFS",
			Decimals:   1,
		}
	}
	d.roomLevel = room("room_level", "Room level", "mdi:waveform")
	d.roomFloor = room("room_floor", "Room floor", "mdi:arrow-collapse-down")

	d.ads = &esphome.Sensor{
		Base: esphome.Base{
			ObjectID: "ble_advertisements", Name: "BLE advertisements", Icon: "mdi:bluetooth-audio",
			Category: esphome.CategoryDiagnostic,
		},
		Unit:       "/s",
		StateClass: esphome.StateClassMeasurement,
		Decimals:   1,
	}

	d.luxPath = metrics.Reader{}.LuxPath()
	switch {
	case d.luxPath != "":
		slog.Info("light sensor", "at", d.luxPath)
	case Ambient != nil:
		slog.Info("light sensor", "at", "alsps")
	default:
		slog.Warn("no light sensor found")
	}
	d.lux = &esphome.Sensor{
		Base: esphome.Base{
			ObjectID: "lux", Name: "Lux", Icon: "mdi:brightness-6",
			Category: esphome.CategoryDiagnostic,
		},
		Unit:        "lx",
		DeviceClass: "illuminance",
		StateClass:  esphome.StateClassMeasurement,
	}
}

func (d *Diag) sample() {
	switch {
	case d.luxPath != "":
		set(d.lux, metrics.Reader{}.Lux(d.luxPath))
	case Ambient != nil:
		set(d.lux, Ambient())
	}

	source := mic.Get()
	d.roomLevel.Set(float32(source.Peak()))
	d.roomFloor.Set(float32(source.Floor()))

	reports := bluetooth.Get().Reports()
	now := time.Now()
	was := d.last
	d.last.at, d.last.reports, d.last.recorded = now, reports, true
	if !was.recorded {
		return
	}
	if secs := now.Sub(was.at).Seconds(); secs > 0 {
		d.ads.Set(float32(float64(reports-was.reports) / secs))
	}
}

func set(s *esphome.Sensor, r metrics.Reading) {
	if r.Known {
		s.Set(float32(r.Value))
	}
}

func inUse() []string { return config.Get().Wake.IDs(wakeword.Slots) }
