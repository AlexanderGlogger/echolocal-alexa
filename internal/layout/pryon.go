package layout

// The Android media bridge and Amazon's native Pryon wake detector, on a biscuit installed with them.
// They live apart from the rest of the layout so the fork that carries them touches nothing shared.
const (
	// AndroidMediaJar is our 32-bit app_process helper. It uses Android's AudioRecord and
	// AudioTrack so Amazon Pryon and EchoLocal share the firmware audio service instead of
	// competing for the raw ALSA capture device.
	AndroidMediaJar = StateDir + "/amazon-helper.jar"
	PryonUIDPath    = StateDir + "/pryon.uid"

	PryonPackage = "com.echolocal.pryon"
	PryonDir     = "/system/priv-app/EchoLocalPryon"
	PryonAPK     = PryonDir + "/EchoLocalPryon.apk"
)
