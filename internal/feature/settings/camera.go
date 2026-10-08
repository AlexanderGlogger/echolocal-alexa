//go:build board_checkers || board_cronos

package settings

import (
	"sync"

	gogui "github.com/go-gui-org/go-gui/gui"
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"

	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/livecam"
	"github.com/ygelfand/echolocal/internal/feature/rtspd"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/setting"
)

var cameraOnce sync.Once
var cameraSettings *sharedsettings.CameraPages[livecam.Knobs]

func cameraPages() *sharedsettings.CameraPages[livecam.Knobs] {
	cameraOnce.Do(func() {
		cameraSettings = sharedsettings.NewCameraPages(
			sharedsettings.CameraOptions[livecam.Knobs]{
				Table:     livecam.Table,
				Read:      livecam.Saved,
				Save:      livecam.Set,
				Reset:     livecam.Reset,
				Stream:    func() bool { return config.Get().RTSP.Enabled },
				SetStream: rtspd.Get().SetEnabled,
				Push:      shell.Get().Push,
				Redraw:    shell.Get().Redraw,
				Preview:   camWant,
				Wrap:      camPage,
			},
		)
	})
	return cameraSettings
}
func cameraPage() *shell.Page                 { return cameraPages().Page() }
func sectionPage(g setting.Group) *shell.Page { return cameraPages().Section(g) }
func choicePage(s livecam.Knob) *shell.Page   { return cameraPages().Choice(s) }
func shown() []setting.Group                  { return cameraPages().Shown() }

func init() {
	sections = append(
		sections,
		section{label: "settings.camera", glyph: gogui.IconCamera, page: cameraPage},
	)
	pages["camera"] = cameraPage
}
