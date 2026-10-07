package speaker

import (
	"github.com/ygelfand/libcountertop/pkg/audio/analysis"
)

// Scale multiplies samples in place, clamping rather than wrapping.
var Scale = analysis.Scale
