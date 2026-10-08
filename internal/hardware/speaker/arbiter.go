package speaker

import "github.com/ygelfand/libcountertop/pkg/audio/background"

type Producer = background.Producer
type Arbiter = background.Arbiter

func (d *Driver) Backgrounds() *Arbiter {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.arb == nil {
		d.arb = &Arbiter{}
		d.bg = d.arb
	}
	return d.arb
}
