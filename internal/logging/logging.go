package logging

import (
	"github.com/ygelfand/libcountertop/pkg/android/logd"

	"github.com/ygelfand/echolocal/internal/layout"
)

var Log = logd.NewLog(layout.LogTag)
