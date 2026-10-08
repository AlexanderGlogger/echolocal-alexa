package theme

// The mark's own colours, sampled from the artwork.
var (
	Blue   = rgb(0x0c9cfc)
	Orange = rgb(0xfc8c0c)

	Ink   = rgb(0x0e0e12)
	Paper = rgb(0xf7f9fb)

	Slate = rgb(0x6b7b8c)

	Mist = rgb(0xe4ebf2)
)

func Brand() Theme {
	return Theme{
		Name: "EchoLocal",
		Dark: false,

		Background: Mist,
		Surface:    Paper,

		Text:  Ink,
		Muted: Slate,

		Accent:  Blue,
		Accent2: Orange,

		Success: rgb(0x2e9e57),
		Warning: rgb(0xb8791a),
		Danger:  rgb(0xd13b3b),
	}
}
