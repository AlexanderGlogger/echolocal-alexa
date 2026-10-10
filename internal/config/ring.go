package config

import "github.com/ygelfand/echolocal/internal/defaults"

type Ring struct {
	Reaction string `json:"reaction"`
	Trouble  string `json:"trouble"`
	Muted    string `json:"muted"`
	Incoming string `json:"incoming"`
	Talking  string `json:"talking"`

	Light Light `json:"light"`
}

type Light struct {
	On         bool    `json:"on"`
	Brightness float32 `json:"brightness"`
	Red        float32 `json:"red"`
	Green      float32 `json:"green"`
	Blue       float32 `json:"blue"`

	Effect string `json:"effect"`
}

const (
	DefaultReaction = ""
	DefaultIncoming = "Beacon"
	DefaultTalking  = "Pulse"

	DefaultBrightness = 1.0
)

func defaultRing(d defaults.Set) Ring {
	return Ring{
		Reaction: DefaultReaction,
		Trouble:  d.RingTrouble,
		Muted:    d.RingMuted,
		Incoming: DefaultIncoming,
		Talking:  DefaultTalking,
		Light:    Light{Brightness: DefaultBrightness, Red: 1, Green: 1, Blue: 1},
	}
}

type RingWriter struct{ st *Store }

func (w RingWriter) Reaction(v string) error {
	return w.st.Update(func(c *Config) { c.Ring.Reaction = v })
}

func (w RingWriter) Trouble(v string) error {
	return w.st.Update(func(c *Config) { c.Ring.Trouble = v })
}

func (w RingWriter) Muted(v string) error {
	return w.st.Update(func(c *Config) { c.Ring.Muted = v })
}

func (w RingWriter) Incoming(v string) error {
	return w.st.Update(func(c *Config) { c.Ring.Incoming = v })
}

func (w RingWriter) Talking(v string) error {
	return w.st.Update(func(c *Config) { c.Ring.Talking = v })
}

func (w RingWriter) Light(v Light) error {
	return w.st.Update(func(c *Config) { c.Ring.Light = v })
}
