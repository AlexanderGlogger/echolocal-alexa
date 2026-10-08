package ui

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"sync"

	"github.com/ygelfand/echolocal/internal/ui/theme"
)

// Scaled from assets by `make logos`.
//
//go:embed logo_light.png
var lightPNG []byte

//go:embed logo_dark.png
var darkPNG []byte

var (
	lightOnce sync.Once
	light     image.Image

	darkOnce sync.Once
	dark     image.Image
)

func Logo() image.Image { return decode(&lightOnce, &light, lightPNG) }

func LogoPNG() []byte { return lightPNG }

func Night() image.Image { return decode(&darkOnce, &dark, darkPNG) }

func Mark(onDark bool) image.Image {
	if onDark {
		return Night()
	}
	return Logo()
}

func MarkWidth(h int) int {
	img := Logo()
	if img == nil || img.Bounds().Dy() == 0 {
		return h
	}
	b := img.Bounds()
	return b.Dx() * h / b.Dy()
}

func DrawLogo(s Surface, r Rect, on theme.Color) {
	img := Logo()
	if theme.Dark(on) {
		img = Night()
	}
	DrawImageFit(s, img, r, on)
}

func decode(once *sync.Once, into *image.Image, raw []byte) image.Image {
	once.Do(func() {
		img, _, err := image.Decode(bytes.NewReader(raw))
		if err == nil {
			*into = img
		}
	})
	return *into
}
