package display

import (
	"github.com/ygelfand/libcountertop/pkg/display/geometry"

	"github.com/ygelfand/echolocal/internal/component"
)

type Orientation = geometry.Orientation

const (
	Rotate0   = geometry.Rotate0
	Rotate90  = geometry.Rotate90
	Rotate180 = geometry.Rotate180
	Rotate270 = geometry.Rotate270
)

func Mounted() Orientation { return Orientation(component.Board().PanelRotation) }
