package ble

import (
	"sync"
	"time"

	"github.com/ygelfand/libcountertop/pkg/bluetooth/hci"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"
	"github.com/ygelfand/libcountertop/pkg/runtime/service"

	"github.com/ygelfand/echolocal/internal/component"
)

const Node = "/dev/stpbt"

func bringUp() (*hci.Port, hci.Version, error) {
	return hci.BringUp(hci.BringUpOptions{Node: Node})
}

func power(bool) (bool, error) { return false, nil }

var Get = sync.OnceValue(func() *hci.Radio {
	return hci.NewRadio(hci.RadioOptions{BringUp: bringUp, Power: power})
})

func init() {
	component.Register(
		sharedcomponent.Hardware,
		Get,
		sharedcomponent.Order(30),
		sharedcomponent.Supervise(service.Restart(2*time.Second, time.Minute)),
	)
}
