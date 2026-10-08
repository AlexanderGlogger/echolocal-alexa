package firmware

import (
	"context"
	"time"

	"github.com/ygelfand/echolocal/internal/update"
)

// ESPHome entities are not polled.
const checkEvery = 24 * time.Hour

const checkSettle = 5 * time.Minute

func (f *Firmware) Run(ctx context.Context) error {
	first := time.NewTimer(checkSettle)
	defer first.Stop()

	select {
	case <-ctx.Done():
		return nil
	case <-first.C:
		update.Commit()
		f.Check(ctx)
	}

	t := time.NewTicker(checkEvery)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			f.Check(ctx)
		}
	}
}
