package vision

import (
	"sync"

	"github.com/ygelfand/libcountertop/pkg/camera/vision"
	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/feature/livecam"
	"github.com/ygelfand/echolocal/internal/hardware/buttons"
)

func init() {
	component.Register(
		sharedcomponent.Device,
		Get,
		sharedcomponent.Order(40),
		component.Needs(board.Camera),
	)
}

var Get = sync.OnceValue(func() *vision.Vision {
	return vision.New(vision.Options{
		DeviceID: component.DeviceCamera,
		Covered:  func() bool { return buttons.Get().Covered() },
		Picture:  livecam.Get().Still,
	})
})
