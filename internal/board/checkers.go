package board

import "github.com/ygelfand/echolocal/internal/layout"

// Checkers is the 1st-generation Echo Show 5.
var Checkers = Board{
	NativeAPI:     25,
	Device:        "checkers",
	Codename:      "checkers",
	Model:         "Echo Show 5 1st gen (checkers)",
	DefaultName:   "Echo Show 5",
	Caps:          Panel | Wifi | Shutter | Camera | MicMutesCamera,
	PanelRotation: 270,
	UISize:        "mini",
	Diagonal:      5.5,
	Tessera:       Tessera{Board: "sunton8048s070", Columns: 4, Rows: 2},
	DecodeWidth:   1920,
	DecodeHeight:  1088,
	SecureDecoders: map[string]string{
		"video/avc":  "OMX.MTK.VIDEO.DECODER.AVC.secure",
		"video/hevc": "OMX.MTK.VIDEO.DECODER.HEVC.secure",
	},

	CameraWidth: 1280, CameraHeight: 720,

	// The mute button reports KEY_POWER, on a node of its own called "gating".
	Keys:    map[uint16]Key{116: Mute},
	Touch:   "goodix-ts",
	MuteDir: "/sys/devices/platform/amazon-gating",

	Service:     layout.Binary,
	ServiceName: layout.Service,
}
