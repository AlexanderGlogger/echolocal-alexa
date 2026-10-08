package reboot

import (
	"sync"

	esphome "github.com/ygelfand/go-esphome-device"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/safe"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/update"
)

func init() {
	component.Register(sharedcomponent.Device, Get, sharedcomponent.Order(90))
}

var (
	once   sync.Once
	shared *Reboot
)

type Reboot struct {
	button *esphome.Button
}

func Get() *Reboot {
	once.Do(func() { shared = build() })
	return shared
}

func build() *Reboot {
	r := &Reboot{}

	r.button = &esphome.Button{
		Base: esphome.Base{
			ObjectID: "restart",
			Name:     "Restart",
			Icon:     "mdi:restart",
			Category: esphome.CategoryDiagnostic,
		},
		DeviceClass: "restart",
		OnPress: func() {
			safe.Go("reboot", func() {
				update.Reboot("pressed in Home Assistant")
			})
		},
	}

	return r
}

func (r *Reboot) Name() string { return "reboot" }

func (r *Reboot) Entities() []esphome.Entity {
	return []esphome.Entity{r.button}
}
