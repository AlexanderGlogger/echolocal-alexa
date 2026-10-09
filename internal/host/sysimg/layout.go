package sysimg

type Layout struct {
	VerityOffset int64

	Patches []Patch

	Adbd BinPatch

	OTAKey, OTAValue string

	PropsPath string
	Props     map[string]string

	Gate []string

	ServiceRC string
}

func (l Layout) Empty() bool {
	return l.VerityOffset == 0 && len(l.Patches) == 0 && l.Adbd.Path == "" &&
		len(l.Props) == 0 && len(l.Gate) == 0 && l.ServiceRC == ""
}

type BinPatch struct {
	Path          string
	Offset        int64
	Before, After []byte
}

var Biscuit = Layout{
	VerityOffset: VerityOffset,
	Patches:      []Patch{DefaultProp, Fstab},
	Adbd:         Adbd,
	OTAKey:       OTAKey,
	OTAValue:     OTAValue,
}

// Checkers and Cronos keep their root filesystem in the boot ramdisk.
var (
	Checkers = Layout{
		OTAKey:   OTAKey,
		OTAValue: OTAValue,
	}
	Cronos = Layout{
		OTAKey:   OTAKey,
		OTAValue: OTAValue,
	}
	Crown = Layout{
		OTAKey:   OTAKey,
		OTAValue: OTAValue,
	}
)

var Rook = Layout{
	PropsPath: "default.prop",
	Props: map[string]string{
		"ro.secure":              "0",
		"ro.adb.secure":          "0",
		"persist.sys.usb.config": "adb",
	},

	Gate:      []string{"surfaceflinger", "audioserver"},
	ServiceRC: "system/etc/init/hw/init.zygote32.rc",
}
