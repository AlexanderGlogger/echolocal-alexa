package face

import (
	shared "github.com/ygelfand/libcountertop/pkg/display/clockface"

	"github.com/ygelfand/echolocal/internal/config"
)

type (
	Face    = shared.Face
	Ticking = shared.Ticking
	Reading = shared.Reading
)

var (
	Read = shared.Read
	Say  = shared.Say
	Lit  = shared.Lit
	Bars = shared.Bars
)

func Of(name config.Face) Face    { return shared.Of(shared.Kind(name)) }
func Ticks(name config.Face) bool { return shared.Ticks(shared.Kind(name)) }
