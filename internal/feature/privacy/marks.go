package privacy

type Marks struct {
	MicMuted      bool
	CameraBlocked bool
}

func (m Marks) Showing() bool { return m.MicMuted || m.CameraBlocked }
