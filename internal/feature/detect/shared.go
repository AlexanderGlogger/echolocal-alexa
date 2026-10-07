package detect

import sharedengine "github.com/ygelfand/libcountertop/pkg/inference/detect"

type Engine = sharedengine.Engine

var New = sharedengine.New

const (
	Hold           = sharedengine.Hold
	Refractory     = sharedengine.Refractory
	NearMiss       = sharedengine.NearMiss
	NearMissSettle = sharedengine.NearMissSettle
)
