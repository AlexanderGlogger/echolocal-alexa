package wake

import (
	"sync"

	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/layout"
)

var get = sync.OnceValue(
	func() *sharedwake.Library { return sharedwake.NewLibrary(layout.ModelDir) },
)

func Lib() *sharedwake.Library { return get() }
