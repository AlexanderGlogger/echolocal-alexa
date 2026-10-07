package bluetooth

import (
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/ble"
	"github.com/ygelfand/echolocal/internal/layout"
	sharedproxy "github.com/ygelfand/libcountertop/pkg/bluetooth/proxy"
	"sync"
)

type Proxy struct{ *sharedproxy.Proxy }

const Features = sharedproxy.Features

func init() { component.Register(component.Device, Get, component.Order(40)) }

var get = sync.OnceValue(func() *Proxy {
	return &Proxy{sharedproxy.New(sharedproxy.Options{
		Radio:     ble.Get(),
		Read:      func() bool { return config.Get().Bluetooth.Proxy },
		Save:      func(v bool) error { return config.Set().Bluetooth().Proxy(v) },
		Reconnect: func() { component.Reconnect.Emit(struct{}{}) },
		Beacon: func() []byte {
			mac, _ := layout.FactoryMAC()
			return sharedproxy.BeaconAdvertisement(sharedproxy.BeaconMinor(mac))
		},
	})}
})

func Get() *Proxy                      { return get() }
func (p *Proxy) Restore(config.Config) { p.Proxy.Restore() }
