package maintenance

import (
	"context"
	"sync"
	"time"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/feature/recording"
)

const Period = 5 * time.Minute

func init() {
	component.Register(component.Device, Get, component.Order(80))
}

var (
	once   sync.Once
	shared *Maintenance
)

type Maintenance struct{}

func Get() *Maintenance {
	once.Do(func() { shared = &Maintenance{} })
	return shared
}

func (m *Maintenance) Name() string { return "maintenance" }

func (m *Maintenance) Run(ctx context.Context) error {
	t := time.NewTicker(Period)
	defer t.Stop()

	for {
		recording.Get().Prune()

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}
