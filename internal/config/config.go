package config

import (
	"github.com/ygelfand/libcountertop/pkg/settings/schema"
	"github.com/ygelfand/libcountertop/pkg/settings/storage"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/defaults"
)

type Config struct {
	Feedback   schema.Feedback  `json:"feedback"`
	Time       schema.Time      `json:"time"`
	Device     Device           `json:"-"`
	Speaker    Speaker          `json:"speaker"`
	Volume     Volume           `json:"volume"`
	Microphone Microphone       `json:"microphone"`
	Wake       schema.Wake      `json:"wake"`
	Ring       Ring             `json:"ring"`
	Update     schema.Update    `json:"update"`
	Bluetooth  schema.Bluetooth `json:"bluetooth"`
	Diag       Diag             `json:"diag"`
	Media      schema.Media     `json:"media"`
	Sendspin   schema.Sendspin  `json:"sendspin"`
	Screen     schema.Screen    `json:"screen"`
	Clock      schema.Clock     `json:"clock"`
	Visual     schema.Visual    `json:"visual"`
	Idle       schema.Idle      `json:"idle"`
	Cast       schema.Cast      `json:"cast"`
	Camera     schema.Camera    `json:"camera"`
	RTSP       schema.RTSP      `json:"rtsp"`
	Poster     schema.Poster    `json:"poster"`
	Network    schema.Network   `json:"network"`
	API        schema.API       `json:"api"`
	Home       schema.Home      `json:"home"`
	Weather    schema.Weather   `json:"weather"`
	Call       schema.Call      `json:"call"`
}

func Defaults() Config {
	d := defaults.Current()

	return Config{
		Feedback:   schema.DefaultFeedback(),
		Network:    schema.DefaultNetwork(),
		Speaker:    defaultSpeaker(),
		Volume:     defaultVolume(),
		Microphone: defaultMicrophone(d),
		Ring:       defaultRing(d),
		Update:     schema.DefaultUpdate(),
		Bluetooth:  schema.DefaultBluetooth(),
		Diag:       defaultDiag(d),
		Media:      schema.DefaultMedia(),
		Sendspin:   schema.DefaultSendspin(),
		Screen:     schema.DefaultScreen(),
		Clock:      schema.DefaultClock(),
		Visual:     schema.DefaultVisualSettings(),
		Idle:       schema.DefaultIdle(),
		Cast:       schema.DefaultCast(),
		Camera:     schema.DefaultCamera(),
		RTSP:       schema.DefaultRTSP(),
		Poster:     schema.DefaultPoster(),
		Home:       schema.DefaultHome(),
		API:        schema.DefaultAPI(),
		Weather:    schema.DefaultWeather(),
		Call:       schema.DefaultCall(),

		Wake: schema.Wake{Stop: schema.DefaultStop()},
	}
}

type Device struct {
	Name string
	Addr string

	Board board.Board
}

type Writer struct{ st *Store }

func (w Writer) Speaker() SpeakerWriter       { return SpeakerWriter(w) }
func (w Writer) Volume() VolumeWriter         { return VolumeWriter(w) }
func (w Writer) Microphone() MicrophoneWriter { return MicrophoneWriter(w) }
func (w Writer) Ring() RingWriter             { return RingWriter(w) }
func (w Writer) Diag() DiagWriter             { return DiagWriter(w) }

func (w Writer) Update() schema.UpdateWriter {
	return schema.NewUpdateWriter(storage.Field(w.st.Store, updateOf))
}

func (w Writer) Bluetooth() schema.BluetoothWriter {
	return schema.NewBluetoothWriter(storage.Field(w.st.Store, bluetoothOf))
}

func (w Writer) Media() schema.MediaWriter {
	return schema.NewMediaWriter(storage.Field(w.st.Store, mediaOf))
}

func (w Writer) Sendspin() schema.SendspinWriter {
	return schema.NewSendspinWriter(storage.Field(w.st.Store, sendspinOf))
}

func (w Writer) Screen() schema.ScreenWriter {
	return schema.NewScreenWriter(storage.Field(w.st.Store, screenOf))
}

func (w Writer) Clock() schema.ClockWriter {
	return schema.NewClockWriter(storage.Field(w.st.Store, clockOf))
}

func (w Writer) Visual() schema.VisualWriter {
	return schema.NewVisualWriter(storage.Field(w.st.Store, visualOf))
}

func (w Writer) Idle() schema.IdleWriter {
	return schema.NewIdleWriter(storage.Field(w.st.Store, idleOf))
}

func (w Writer) Cast() schema.CastWriter {
	return schema.NewCastWriter(storage.Field(w.st.Store, castOf))
}

func (w Writer) Camera() schema.CameraWriter {
	return schema.NewCameraWriter(storage.Field(w.st.Store, cameraOf))
}

func (w Writer) RTSP() schema.RTSPWriter {
	return schema.NewRTSPWriter(storage.Field(w.st.Store, rtspOf))
}

func (w Writer) Poster() schema.PosterWriter {
	return schema.NewPosterWriter(storage.Field(w.st.Store, posterOf))
}

func (w Writer) Network() schema.NetworkWriter {
	return schema.NewNetworkWriter(storage.Field(w.st.Store, networkOf))
}

func (w Writer) API() schema.APIWriter {
	return schema.NewAPIWriter(storage.Field(w.st.Store, apiOf))
}

func (w Writer) Home() schema.HomeWriter {
	return schema.NewHomeWriter(storage.Field(w.st.Store, homeOf))
}

func (w Writer) Weather() schema.WeatherWriter {
	return schema.NewWeatherWriter(storage.Field(w.st.Store, weatherOf))
}

func (w Writer) Call() schema.CallWriter {
	return schema.NewCallWriter(storage.Field(w.st.Store, callOf))
}

func (w Writer) Time() schema.TimeWriter {
	return schema.NewTimeWriter(storage.Field(w.st.Store, timeOf))
}

func (w Writer) Wake(slot int) schema.WakeWriter {
	return schema.NewWakeWriter(storage.Field(w.st.Store, wakeOf), slot)
}

func (w Writer) Stop() schema.StopWriter {
	return schema.NewStopWriter(storage.Field(w.st.Store, wakeOf))
}

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
