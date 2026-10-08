package config

type Update struct {
	Channel string `json:"channel"`

	LastVersion string `json:"last_version"`
}

const DefaultChannel = "stable"

func defaultUpdate() Update {
	return Update{Channel: DefaultChannel}
}

type UpdateWriter struct{ st *Store }

func (w UpdateWriter) Channel(v string) error {
	return w.st.Update(func(c *Config) { c.Update.Channel = v })
}

func (w UpdateWriter) LastVersion(v string) error {
	return w.st.Update(func(c *Config) { c.Update.LastVersion = v })
}
