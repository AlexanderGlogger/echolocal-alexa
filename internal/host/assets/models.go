package assets

import (
	"embed"
	"io/fs"
)

// okay_nabu, hey_jarvis and hey_mycroft, from esphome/micro-wake-word-models.
//
//go:embed models
var models embed.FS

type Model struct {
	ID       string
	Model    []byte
	Manifest []byte
}

func Models() ([]Model, error) {
	entries, err := fs.ReadDir(models, "models")
	if err != nil {
		return nil, err
	}

	var out []Model
	for _, e := range entries {
		name := e.Name()
		if len(name) < 8 || name[len(name)-7:] != ".tflite" {
			continue
		}
		id := name[:len(name)-7]

		model, err := models.ReadFile("models/" + name)
		if err != nil {
			return nil, err
		}
		manifest, err := models.ReadFile("models/" + id + ".json")
		if err != nil {
			return nil, err
		}
		out = append(out, Model{ID: id, Model: model, Manifest: manifest})
	}
	return out, nil
}
