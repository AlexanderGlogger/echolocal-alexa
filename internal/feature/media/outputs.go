package media

import (
	"github.com/ygelfand/libcountertop/pkg/say"
	setting "github.com/ygelfand/libcountertop/pkg/settings"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/hardware/speaker"
)

func outputControls() *setting.Controls[config.Speaker] {
	if !component.Board().Has(board.HeadphoneJack) {
		return nil
	}
	return &setting.Controls[config.Speaker]{
		Table:  outputTable(),
		Device: component.DevicePlayback,
		Read:   config.SpeakerSection.Read,
		Save: func(s setting.Setting[config.Speaker], v string) error {
			if err := setting.Store(config.SpeakerSection, s, v); err != nil {
				return err
			}
			c := config.SpeakerSection.Read()
			speaker.Get().SetOutputs(c.SpeakerOutput, c.HeadphoneOutput)
			return nil
		},
	}
}

func outputTable() *setting.Table[config.Speaker] {
	modes := config.OutputModes()
	return setting.NewTable(
		"speaker",
		[]setting.Group{component.SpeakerGroup},
		[]setting.Setting[config.Speaker]{
			setting.Pick(
				setting.Setting[config.Speaker]{
					Name:  "speaker_output",
					ID:    "speaker_output",
					Group: component.SpeakerGroup,
					Icon:  "mdi:speaker",
				},
				modes,
				func(c *config.Speaker) *config.OutputMode { return &c.SpeakerOutput },
			),
			setting.Pick(
				setting.Setting[config.Speaker]{
					Name:  "headphone_output",
					ID:    "headphone_output",
					Group: component.SpeakerGroup,
					Icon:  "mdi:headphones",
				},
				modes,
				func(c *config.Speaker) *config.OutputMode { return &c.HeadphoneOutput },
			),
		},
		setting.Messages{Text: say.T, Missing: say.Missing},
	)
}
