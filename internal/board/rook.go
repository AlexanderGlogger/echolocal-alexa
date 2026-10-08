package board

// Rook is the 1st-generation Echo Spot, running LineageOS.
var Rook = Board{
	Device:       "rook",
	Codename:     "rook",
	SystemAsRoot: true,
	Model:        "Echo Spot (rook)",
	DefaultName:  "Echo Spot",
	Caps:         Panel | Wifi,
	TesseraBoard: "guition",

	Service:     "/system/app/echod/echod",
	ServiceName: "zygote",
}
