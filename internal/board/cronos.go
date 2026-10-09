package board

import "github.com/ygelfand/echolocal/internal/layout"

// Cronos is the 2nd-generation Echo Show 5.
var Cronos = Board{
	NativeAPI:     25,
	Device:        "cronos",
	Codename:      "cronos",
	Model:         "Echo Show 5 2nd gen (cronos)",
	DefaultName:   "Echo Show 5",
	Caps:          Panel | Wifi | Shutter | Camera | MicMutesCamera,
	PanelRotation: 270,
	UISize:        "mini",
	Diagonal:      5.5,
	Tessera:       Tessera{Board: "sunton8048s070", Columns: 4, Rows: 2},
	SecureDecoders: map[string]string{
		"video/avc":  "OMX.MTK.VIDEO.DECODER.AVC.secure",
		"video/hevc": "OMX.MTK.VIDEO.DECODER.HEVC.secure",
	},
	Keys:    map[uint16]Key{116: Mute},
	MuteDir: "/sys/devices/platform/gpio-privacy",

	CameraWidth: 1280, CameraHeight: 720,

	Service:     layout.Binary,
	ServiceName: layout.Service,
}
