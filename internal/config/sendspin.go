package config

import "github.com/ygelfand/libcountertop/pkg/settings/schema"

// Sendspin is the room's part of a synchronized stream. A server discovers it and dials in, so there
// is no address to set.
type Sendspin = schema.Sendspin

var defaultSendspin = schema.DefaultSendspin

type SendspinWriter struct{ st *Store }

func (w SendspinWriter) Enabled(v bool) error {
	return w.st.Update(func(c *Config) { c.Sendspin.Enabled = v })
}
