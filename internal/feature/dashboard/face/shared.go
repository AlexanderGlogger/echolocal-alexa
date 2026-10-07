// Package face adapts shared clock faces to product settings.
package face

import (
	"github.com/ygelfand/echolocal/internal/config"
	shared "github.com/ygelfand/libcountertop/pkg/display/clockface"
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
