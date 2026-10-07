package gpu

import (
	"github.com/ygelfand/echolocal/internal/hardware/display"
	sharedgpu "github.com/ygelfand/libcountertop/pkg/display/gpu"
)

var renderer = sharedgpu.New(display.Get())

func Get() *sharedgpu.Renderer { return renderer }
