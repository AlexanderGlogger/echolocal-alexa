package theme

import sharedtheme "github.com/ygelfand/libcountertop/pkg/display/theme"

// The mark's own colours, sampled from the artwork.
var (
	Blue   = sharedtheme.RGB(0x0c9cfc)
	Orange = sharedtheme.RGB(0xfc8c0c)

	Ink   = sharedtheme.RGB(0x0e0e12)
	Paper = sharedtheme.RGB(0xf7f9fb)

	Slate = sharedtheme.RGB(0x6b7b8c)

	Mist = sharedtheme.RGB(0xe4ebf2)
)

func Brand() sharedtheme.Theme {
	return sharedtheme.Theme{
		Name: "EchoLocal",
		Dark: false,

		Background: Mist,
		Surface:    Paper,

		Text:  Ink,
		Muted: Slate,

		Accent:  Blue,
		Accent2: Orange,

		Success: sharedtheme.RGB(0x2e9e57),
		Warning: sharedtheme.RGB(0xb8791a),
		Danger:  sharedtheme.RGB(0xd13b3b),
	}
}
