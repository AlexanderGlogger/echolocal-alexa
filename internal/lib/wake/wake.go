package wake

import (
	"sync"

	sharedwake "github.com/ygelfand/libcountertop/pkg/inference/wake"

	"github.com/ygelfand/echolocal/internal/layout"
)

var get = sync.OnceValue(
	func() *Library { return withPryon(sharedwake.NewLibrary(layout.ModelDir)) },
)

func Lib() *Library { return get() }
