//go:build board_crown

package mic

// Device 22 offers 16 kHz, S24_3LE, 6 channels.
const (
	Channels      = 6
	CaptureDevice = 22
)

const Mics = 4

const CenterMic = 0

var adcs = []string{"A", "B"}

const beamforming = false

const lowCutHz = 100
