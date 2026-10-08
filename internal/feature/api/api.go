package api

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/libcountertop/pkg/runtime/safe"

	"github.com/ygelfand/echolocal/internal/android/firewall"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/bluetooth"
	"github.com/ygelfand/echolocal/internal/feature/voice"
	"github.com/ygelfand/echolocal/internal/feature/wakeword"
	"github.com/ygelfand/echolocal/internal/layout"
)

func init() {
	component.Register(component.Network, Get, component.Order(99))
}

// Home Assistant drops the connection after 4.5 keepalive intervals.
const writeTimeout = 20 * time.Second

type API struct {
	srv  *esphome.Server
	name string

	reconnect chan struct{}
	announced sync.Once

	listening atomic.Bool
}

var (
	once   sync.Once
	shared *API
)

func Get() *API {
	once.Do(func() {
		shared = &API{reconnect: make(chan struct{}, 1)}
		component.Reconnect.Listen(func(struct{}) { shared.Reconnect() })
		component.Fire.Listen(func(e component.Event) { shared.fire(e) })
	})
	return shared
}

func (a *API) Name() string { return "api" }

func (a *API) Startup() component.Progress {
	if !a.listening.Load() {
		return component.Progress{Doing: "opening the port"}
	}
	return component.Progress{Done: true, Doing: fmt.Sprintf("port %d", layout.Port)}
}

func (a *API) Start(context.Context) error {
	psk, err := loadPSK(layout.KeyPath)
	if err != nil {
		return err
	}
	mac, err := layout.FactoryMAC()
	if err != nil {
		return err
	}

	device := config.Get().Device
	registered := component.Default().Entities()
	ents := esphome.NewEntities()
	if err := ents.Add(registered...); err != nil {
		return err
	}
	if err := ents.AddActions(component.Default().Actions()...); err != nil {
		return err
	}

	a.name = layout.Slug(device.Name)
	a.srv = &esphome.Server{
		Addr:         device.Addr,
		WriteTimeout: writeTimeout,
		Info: esphome.Info{
			Name:              a.name,
			FriendlyName:      device.Name,
			MACAddress:        mac,
			Manufacturer:      layout.Manufacturer,
			Model:             device.Board.Model,
			Version:           layout.Version,
			ProjectName:       layout.Manufacturer + "." + device.Board.Codename,
			VoiceFeatures:     voice.Features,
			BluetoothFeatures: bluetooth.Get().Features(),

			Devices: subDevices(device.Name, registered),
		},
		PSK:                psk,
		Logger:             slog.Default(),
		OnSetEncryptionKey: func(k esphome.PSK) error { return writePSK(layout.KeyPath, k) },

		OnSubscribed: func() {
			adopted()
			component.Subscribed.Emit(struct{}{})
		},

		Handler: esphome.Chain(
			append([]esphome.Handler{ents}, component.Default().Handlers()...)...),
	}
	return nil
}

func subDevices(name string, ents []esphome.Entity) []esphome.Device {
	all := []esphome.Device{
		{ID: component.DeviceRing, Name: name + " ring"},
		{ID: component.DeviceMicrophone, Name: name + " microphone"},
		{ID: component.DevicePlayback, Name: name + " playback"},
		{ID: component.DeviceScreen, Name: name + " screen"},
		{ID: component.DeviceCamera, Name: name + " camera"},
	}
	for slot := range wakeword.Slots {
		all = append(all, esphome.Device{
			ID:   component.AssistantDevice(slot),
			Name: fmt.Sprintf("%s assistant %d", name, slot+1),
		})
	}

	used := map[uint32]bool{}
	for _, e := range ents {
		used[deviceOf(e)] = true
	}

	var out []esphome.Device
	for _, d := range all {
		if used[d.ID] {
			out = append(out, d)
		}
	}
	return out
}

func deviceOf(e esphome.Entity) uint32 {
	v := reflect.ValueOf(e)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0
	}
	if f := v.FieldByName("DeviceID"); f.IsValid() && f.Kind() == reflect.Uint32 {
		return uint32(f.Uint())
	}
	return 0
}

func (a *API) Run(ctx context.Context) error {
	defer a.listening.Store(false)
	safe.Go("logs", func() { a.pipeLogs(ctx) })

	if err := firewall.Open(firewall.API, layout.Port); err != nil {
		slog.Error("opening the api port failed", "port", layout.Port, "err", err)
	}

	for {
		ln, err := net.Listen("tcp", a.srv.Addr)
		if err != nil {
			return fmt.Errorf("api: listen %s: %w", a.srv.Addr, err)
		}

		a.listening.Store(true)

		a.announced.Do(func() {
			safe.Go("mdns", func() { a.advertise(ctx, ln.Addr().(*net.TCPAddr).Port) })
		})

		serving, stop := context.WithCancel(ctx)
		go func() {
			select {
			case <-a.reconnect:
			case <-serving.Done():
			}
			stop()
		}()

		err = a.srv.Serve(serving, ln)
		stop()
		a.listening.Store(false)

		if err != nil || ctx.Err() != nil {
			return err
		}

		a.srv.Info.BluetoothFeatures = bluetooth.Get().Features()
		slog.Info("serving again", "bluetooth", a.srv.Info.BluetoothFeatures)
	}
}

func (a *API) fire(e component.Event) {
	if a.srv == nil {
		return
	}
	if err := a.srv.FireEvent(e.Name, e.Data); err != nil {
		slog.Debug("firing an event failed", "event", e.Name, "err", err)
	}
}

func (a *API) Reconnect() {
	select {
	case a.reconnect <- struct{}{}:
	default:
	}
}

func loadPSK(path string) (*esphome.PSK, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return esphome.Unprovisioned(), nil
	}
	k, err := esphome.ParsePSK(strings.TrimSpace(string(b)))
	if err != nil {
		return nil, fmt.Errorf("api: key at %s: %w", path, err)
	}
	return &k, nil
}

func Adopted() bool { return config.Get().API.Adopted }

func Key() string {
	b, err := os.ReadFile(layout.KeyPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func adopted() {
	if Adopted() {
		return
	}

	if err := config.Set().API().Adopted(true); err != nil {
		slog.Error("recording adoption failed", "err", err)
		return
	}
	slog.Info("adopted")
}

func writePSK(path string, k esphome.PSK) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(k.String()+"\n"), 0o600)
}
