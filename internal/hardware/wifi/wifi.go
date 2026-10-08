package wifi

import (
	"sync"
	"time"

	sharedwifi "github.com/ygelfand/libcountertop/pkg/network/wifi"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/service"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/layout"
)

func init() {
	component.Register(sharedcomponent.Hardware, Get, sharedcomponent.Order(20),
		component.Needs(board.Wifi),
		sharedcomponent.Supervise(service.Restart(2*time.Second, time.Minute)))
}

var get = sync.OnceValue(func() *sharedwifi.Radio {
	return sharedwifi.New(sharedwifi.Options{
		Interface:  layout.WifiIface,
		SocketDir:  layout.WifiSockets,
		ConfigPath: layout.WifiConf,
		Bgscan:     layout.WifiBgscan,
		User:       layout.WifiUser,
		Client:     "echod",
	})
})

func Get() *sharedwifi.Radio { return get() }
