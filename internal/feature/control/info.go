package control

import (
	"time"

	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/metrics"
	"github.com/ygelfand/echolocal/internal/layout"
)

func deviceInfo() harness.DeviceInfo {
	d := config.Get().Device
	return harness.DeviceInfo{
		Version: layout.Version,
		Commit:  layout.GitCommit,
		Built:   layout.BuildDate,
		Board:   component.Board().Codename,
		Name:    d.Name,
		Address: d.Addr,
		Model:   d.Board.Model,
		Uptime:  time.Duration(metrics.Uptime() * float64(time.Second)),
	}
}
