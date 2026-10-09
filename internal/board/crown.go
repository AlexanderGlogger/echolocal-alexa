package board

import "github.com/ygelfand/echolocal/internal/layout"

// Crown is the 1st-generation Echo Show 8.
var Crown = Board{
	NativeAPI:     25,
	Device:        "crown",
	Codename:      "crown",
	Model:         "Echo Show 8 1st gen (crown)",
	DefaultName:   "Echo Show 8",
	Caps:          Panel | Wifi | Shutter | Camera | MicMutesCamera,
	PanelRotation: 270,
	UISize:        "mini",
	Diagonal:      8.0,
	Tessera:       Tessera{Board: "sunton8048s070", Columns: 4, Rows: 2},
	SecureDecoders: map[string]string{
		"video/avc":  "OMX.MTK.VIDEO.DECODER.AVC.secure",
		"video/hevc": "OMX.MTK.VIDEO.DECODER.HEVC.secure",
	},
	Keys:    map[uint16]Key{116: Mute},
	Touch:   "fts_ts",
	MuteDir: "/sys/devices/platform/amazon-gating",

	CameraWidth: 1280, CameraHeight: 720,

	Service:     layout.Binary,
	ServiceName: layout.Service,
}
