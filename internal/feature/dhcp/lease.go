package dhcp

import (
	"github.com/ygelfand/echolocal/internal/config"
	"github.com/ygelfand/echolocal/internal/layout"
)

const fallback = "echolocal"

func hostname() string {
	if slug := layout.Slug(config.Get().Device.Name); slug != "" {
		return slug
	}
	return fallback
}

type saved struct{}

func (saved) Address() string              { return config.Get().Network.Address }
func (saved) SetAddress(said string) error { return config.Set().Network().Address(said) }
