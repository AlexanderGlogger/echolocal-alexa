package speaker

import (
	"sync"

	"github.com/ygelfand/libcountertop/pkg/audio/sound"
)

type Driver = sound.Driver[*Player]

var Sound = sync.OnceValue(func() *Driver { return sound.NewDriver(Get()) })
