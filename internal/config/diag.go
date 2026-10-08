package config

import "github.com/ygelfand/echolocal/internal/defaults"

type Diag struct {
	Interval int `json:"interval"`

	ADB bool `json:"remote_adb"`

	MinCores int `json:"min_cores"`
}

const DefaultInterval = 300

func defaultDiag(d defaults.Set) Diag {
	return Diag{Interval: DefaultInterval, MinCores: d.MinCores}
}

type DiagWriter struct{ st *Store }

func (w DiagWriter) Interval(v int) error {
	return w.st.Update(func(c *Config) { c.Diag.Interval = v })
}

func (w DiagWriter) ADB(v bool) error {
	return w.st.Update(func(c *Config) { c.Diag.ADB = v })
}

func (w DiagWriter) MinCores(v int) error {
	return w.st.Update(func(c *Config) { c.Diag.MinCores = v })
}
