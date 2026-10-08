//go:build board_checkers || board_cronos

package settings

import (
	sharedpreview "github.com/ygelfand/libcountertop/pkg/display/camerapreview"

	"github.com/ygelfand/echolocal/internal/feature/livecam"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/hardware/display"
)

var cameraPreview = sharedpreview.New(
	sharedpreview.Dependencies{
		Shell:   shell.Get(),
		Camera:  livecam.Sessions(),
		Display: display.Get(),
	},
)
var camPage = cameraPreview.Page
var camWant = cameraPreview.Want
