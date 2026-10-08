package rtspd

import (
	"log/slog"
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	"github.com/ygelfand/libcountertop/pkg/camera/live"
	"github.com/ygelfand/libcountertop/pkg/camera/onvif"
	camerartspd "github.com/ygelfand/libcountertop/pkg/camera/rtspd"
	"github.com/ygelfand/libcountertop/pkg/media/rtsp"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/android/prop"
	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/livecam"
	"github.com/ygelfand/echolocal/internal/feature/vision"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
	"github.com/ygelfand/echolocal/internal/layout"
)

const (
	Port      = 8554
	ONVIFPort = 8000
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(70),
		component.Needs(board.Camera),
	)
}

type Server struct {
	sw  *esphome.Switch
	srv *camerartspd.Server
}

var (
	once   sync.Once
	shared *Server
)

func Get() *Server {
	once.Do(func() {
		shared = &Server{srv: camerartspd.New(camerartspd.Options{
			Port:     Port,
			Identity: rtsp.Identity{Name: "EchoLocal", CNAMEPrefix: "echolocal-"},
			Source:   livecam.Get().Sessions(),
			Helper:   "echolocal-camera",
			Served:   livecam.Get().Served,
			FPS:      live.FPS,
			Bitrate:  livecam.Get().Bitrate,
			Allowed:  func() bool { return !buttons.Get().Covered() },
			ONVIF:    describe,
		})}
		shared.build()
		livecam.Get().StreamsChanged.Listen(func([]int) { shared.reload() })
	})
	return shared
}

func describe() onvif.Config {
	dev := config.Get().Device
	serial, _ := prop.Get("ro.serialno")
	if serial == "" {
		serial = dev.Name
	}
	return onvif.Config{
		Port: ONVIFPort,
		Device: onvif.Device{
			Manufacturer: layout.Manufacturer,
			Model:        dev.Board.Model,
			Firmware:     layout.Version,
			Serial:       serial,
			Hardware:     dev.Board.Model,
			Name:         dev.Name,
		},
		Hardware: "EchoLocal",
		Snapshot: vision.Get().Still,
	}
}

func (s *Server) Name() string { return "rtsp" }

func (s *Server) Entities() []esphome.Entity { return []esphome.Entity{s.sw} }

func (s *Server) Restore(cfg config.Config) {
	s.sw.Set(cfg.RTSP.Enabled)
	if cfg.RTSP.Enabled {
		s.srv.Start()
	}
}

func (s *Server) build() {
	s.sw = &esphome.Switch{
		Base: esphome.Base{ObjectID: "rtsp", Name: "RTSP server", Icon: "mdi:cctv",
			Category: esphome.CategoryConfig, DeviceID: component.DeviceCamera},
	}
	s.sw.OnCommand = s.SetEnabled
}

func (s *Server) SetEnabled(on bool) {
	s.sw.Set(on)
	if err := config.Set().RTSP().Enabled(on); err != nil {
		slog.Error("saving a setting failed", "setting", s.sw.ObjectID, "err", err)
	}
	if on {
		s.srv.Start()
	} else {
		s.srv.Stop()
	}
}

func (s *Server) reload() {
	if !config.Get().RTSP.Enabled {
		return
	}
	s.srv.Stop()
	s.srv.Start()
}
