package bluetooth

import (
	"sync"

	sharedproxy "github.com/ygelfand/libcountertop/pkg/bluetooth/proxy"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/ble"
	"github.com/ygelfand/echolocal/internal/layout"
)

type Proxy struct{ *sharedproxy.Proxy }

const Features = sharedproxy.Features

func init() { component.Register(component.Device, Get, component.Order(40)) }

var get = sync.OnceValue(func() *Proxy {
	return &Proxy{sharedproxy.New(sharedproxy.Dependencies{
		Radio:     ble.Get(),
		Settings:  config.BluetoothSection,
		Reconnect: func() { component.Reconnect.Emit(struct{}{}) },
		Beacon: func() []byte {
			mac, _ := layout.FactoryMAC()
			return sharedproxy.BeaconAdvertisement(sharedproxy.BeaconMinor(mac))
		},
	})}
})

func Get() *Proxy                      { return get() }
func (p *Proxy) Restore(config.Config) { p.Proxy.Restore() }
