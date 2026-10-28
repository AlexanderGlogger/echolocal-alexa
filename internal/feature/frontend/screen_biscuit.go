//go:build board_biscuit

package frontend

import (
	"image"
	"net/http"

	"github.com/ygelfand/libcountertop/pkg/input/textentry"
)

var shot func() (image.Image, error)

func addDesktop(map[string]http.Handler, *textentry.Controller) {}
