package component

// Home Assistant keys a sub-device on the device address and this number.
const (
	DeviceRing uint32 = iota + 1

	DeviceMicrophone

	DevicePlayback

	DeviceAssistant
)

const DeviceScreen uint32 = 100

const DeviceCamera uint32 = 101

func AssistantDevice(slot int) uint32 { return DeviceAssistant + uint32(slot) }
