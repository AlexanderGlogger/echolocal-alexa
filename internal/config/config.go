package config

import (
	"fmt"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/defaults"
)

type Config struct {
	Feedback   Feedback   `json:"feedback"`
	Time       Time       `json:"time"`
	Device     Device     `json:"-"`
	Speaker    Speaker    `json:"speaker"`
	Volume     Volume     `json:"volume"`
	Microphone Microphone `json:"microphone"`
	Wake       Wake       `json:"wake"`
	Ring       Ring       `json:"ring"`
	Update     Update     `json:"update"`
	Bluetooth  Bluetooth  `json:"bluetooth"`
	Diag       Diag       `json:"diag"`
	Media      Media      `json:"media"`
	Sendspin   Sendspin   `json:"sendspin"`
	Screen     Screen     `json:"screen"`
	Clock      Clock      `json:"clock"`
	Visual     Visual     `json:"visual"`
	Idle       Idle       `json:"idle"`
	Cast       Cast       `json:"cast"`
	Camera     Camera     `json:"camera"`
	RTSP       RTSP       `json:"rtsp"`
	Poster     Poster     `json:"poster"`
	Network    Network    `json:"network"`
	API        API        `json:"api"`
	Home       Home       `json:"home"`
	Weather    Weather    `json:"weather"`
	Call       Call       `json:"call"`
}

func Defaults() Config {
	d := defaults.Current()

	return Config{
		Feedback:   defaultFeedback(),
		Network:    defaultNetwork(),
		Speaker:    defaultSpeaker(),
		Volume:     defaultVolume(),
		Microphone: defaultMicrophone(d),
		Ring:       defaultRing(d),
		Update:     defaultUpdate(),
		Bluetooth:  defaultBluetooth(),
		Diag:       defaultDiag(d),
		Media:      defaultMedia(),
		Sendspin:   defaultSendspin(),
		Screen:     defaultScreen(),
		Clock:      defaultClock(),
		Visual:     defaultVisual(),
		Idle:       defaultIdle(),
		Cast:       defaultCast(),
		Camera:     defaultCamera(),
		RTSP:       defaultRTSP(),
		Poster:     defaultPoster(),
		Home:       defaultHome(),
		API:        defaultAPI(),
		Weather:    defaultWeather(),
		Call:       defaultCall(),

		Wake: Wake{Stop: defaultStop()},
	}
}

type Device struct {
	Name string
	Addr string

	Board board.Board
}

type Writer struct{ st *Store }

func (w Writer) Time() TimeWriter             { return TimeWriter(w) }
func (w Writer) Speaker() SpeakerWriter       { return SpeakerWriter(w) }
func (w Writer) Volume() VolumeWriter         { return VolumeWriter(w) }
func (w Writer) Microphone() MicrophoneWriter { return MicrophoneWriter(w) }
func (w Writer) Ring() RingWriter             { return RingWriter(w) }
func (w Writer) Update() UpdateWriter         { return UpdateWriter(w) }
func (w Writer) Bluetooth() BluetoothWriter   { return BluetoothWriter(w) }
func (w Writer) Diag() DiagWriter             { return DiagWriter(w) }
func (w Writer) Media() MediaWriter           { return MediaWriter(w) }
func (w Writer) Sendspin() SendspinWriter     { return SendspinWriter(w) }
func (w Writer) Screen() ScreenWriter         { return ScreenWriter(w) }
func (w Writer) Clock() ClockWriter           { return ClockWriter(w) }
func (w Writer) Visual() VisualWriter         { return VisualWriter(w) }
func (w Writer) Idle() IdleWriter             { return IdleWriter(w) }
func (w Writer) Cast() CastWriter             { return CastWriter(w) }
func (w Writer) Camera() CameraWriter         { return CameraWriter(w) }
func (w Writer) RTSP() RTSPWriter             { return RTSPWriter(w) }
func (w Writer) Poster() PosterWriter         { return PosterWriter(w) }
func (w Writer) Network() NetworkWriter       { return NetworkWriter(w) }
func (w Writer) API() APIWriter               { return APIWriter(w) }
func (w Writer) Home() HomeWriter             { return HomeWriter(w) }
func (w Writer) Weather() WeatherWriter       { return WeatherWriter(w) }
func (w Writer) Call() CallWriter             { return CallWriter(w) }

func (w Writer) Wake(slot int) WakeWriter { return WakeWriter{st: w.st, slot: slot} }

func (w Writer) Stop() StopWriter { return StopWriter(w) }

func errSlot(n int) error { return fmt.Errorf("config: wake slot %d", n) }

type Labelled interface{ Label() string }

func Labels[T Labelled](values []T) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, v.Label())
	}
	return out
}

func ByLabel[T Labelled](values []T, label string) (T, bool) {
	for _, v := range values {
		if v.Label() == label {
			return v, true
		}
	}
	var zero T
	return zero, false
}
