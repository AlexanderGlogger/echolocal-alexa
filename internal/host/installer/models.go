package installer

import (
	"fmt"
	"path"

	"github.com/ygelfand/echolocal/internal/host/assets"
	"github.com/ygelfand/echolocal/internal/layout"
)

func installModels(r *run) (string, bool, error) {
	models, err := assets.Models()
	if err != nil {
		return "", false, err
	}
	if len(models) == 0 {
		return "this build ships none", true, nil
	}

	if _, err := r.d.Shell("mkdir -p " + layout.ModelDir); err != nil {
		return "", false, err
	}

	var written, kept int
	for _, m := range models {
		modelPath := path.Join(layout.ModelDir, m.ID+".tflite")

		have, err := r.d.Exists(modelPath)
		if err != nil {
			return "", false, err
		}
		if have {
			kept++
			continue
		}

		if err := r.d.WriteFile(
			path.Join(layout.ModelDir, m.ID+".json"),
			m.Manifest,
			0o644,
		); err != nil {
			return "", false, err
		}
		if err := r.d.WriteFile(modelPath, m.Model, 0o644); err != nil {
			return "", false, err
		}
		written++
	}

	if written == 0 {
		return fmt.Sprintf("%d already there", kept), true, nil
	}
	return fmt.Sprintf("%d written, %d already there", written, kept), false, nil
}
