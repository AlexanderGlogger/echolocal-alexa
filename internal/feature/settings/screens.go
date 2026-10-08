//go:build board_checkers || board_cronos || board_rook

package settings

import (
	sharedvolume "github.com/ygelfand/libcountertop/pkg/audio/volume"
	sharedsettings "github.com/ygelfand/libcountertop/pkg/display/settings"
	"github.com/ygelfand/libcountertop/pkg/say"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/feature/dhcp"
	"github.com/ygelfand/echolocal/internal/feature/homecontrol"
	"github.com/ygelfand/echolocal/internal/feature/media"
	"github.com/ygelfand/echolocal/internal/feature/network"
	"github.com/ygelfand/echolocal/internal/feature/shell"
	"github.com/ygelfand/echolocal/internal/feature/visuals"
	"github.com/ygelfand/echolocal/internal/feature/volume"
	"github.com/ygelfand/echolocal/internal/hardware/wifi"
	"github.com/ygelfand/echolocal/internal/layout"
	"github.com/ygelfand/echolocal/internal/ui/visual"
)

func root() *shell.Page {
	return sharedsettings.RootPage(sharedsettings.RootDependencies{
		Network:      wifi.Get(),
		Screen:       config.ScreenSection,
		Basic:        basicPages(),
		Assistants:   assistantPages(),
		Home:         homecontrol.Get(),
		Shell:        shell.Get(),
		Version:      layout.Version,
		Volume:       sharedvolume.Steps(media.Get(), media.VolumeSteps),
		NetworkPage:  func() shell.View { return networkPage() },
		FeaturesPage: func() shell.View { return featuresPage() },
		VolumePage:   func() shell.View { return volume.Page() },
		AboutPage:    func() shell.View { return aboutPage() },
		Sections: func() []sharedsettings.Section {
			out := make([]sharedsettings.Section, 0, len(sections))
			for _, s := range sections {
				out = append(
					out,
					sharedsettings.Section{
						Label: s.label,
						Glyph: s.glyph,
						Page:  func() shell.View { return s.page() },
					},
				)
			}
			return out
		},
	})
}

type section struct {
	label string
	glyph string
	page  func() *shell.Page
}

var sections []section

var fpsIndex = sharedsettings.FpsIndex

var fpsLevel = sharedsettings.FpsLevel

var fpsSnap = sharedsettings.FpsSnap

var seedOf = sharedsettings.SeedOf

var seedLevel = sharedsettings.SeedLevel

func visualPage() *shell.Page {
	return visualPicker(say.T("debug.visuals"), "", nil, nil, func(k visual.Kind) {
		visuals.Get().SetKind(k)
		shell.Get().Push(visuals.Get().View())
	})
}

func visualPicker(
	title, empty string,
	none func(),
	chosen func() string,
	pick func(visual.Kind),
) *shell.Page {
	return sharedsettings.VisualPicker(title, empty, none, chosen, pick, visual.ThumbnailFit)
}

func uiSize() string {
	if s := config.Get().Screen.Size; s != "" {
		return s
	}
	return component.Board().UISize
}

func networkPage() *shell.Page {
	return sharedsettings.NetworkPage(func() sharedsettings.NetworkState {
		c := sharedsettings.NetworkState{
			SSID:    wifi.Get().Network(),
			Address: address(),
			MAC:     wifi.Get().MAC(),
			Verify:  config.Get().Network.Verify,
		}
		if l := dhcp.Get().Lease(); l != nil {
			c.Router = l.Router.String()
			c.Renew = l.Renew
		}
		return c
	}, network.Get().SetVerify)
}

func aboutPage() *shell.Page {
	return sharedsettings.AboutPage(func() sharedsettings.About {
		return sharedsettings.About{
			Name:    config.Get().Device.Name,
			Version: layout.Version,
			Built:   layout.BuildDate,
			Commit:  layout.GitCommit,
		}
	})
}

func address() string {
	l := dhcp.Get().Lease()
	if l == nil {
		return say.T("network.address.none")
	}
	return l.Address.IP.String()
}

// Below 0.5 a model fires on anything; at 1 nothing reaches it.
const (
	leastSure = 50
	mostSure  = 99
)

var sensitivity = sharedsettings.Sensitivity

func setThreshold(slot int) func(int) { return assistantPages().Threshold(slot) }
