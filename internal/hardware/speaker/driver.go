package speaker

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

type Driver struct {
	p *Player

	mu  sync.Mutex
	now *Claim
	bg  Background
	arb *Arbiter

	yielded bool
}

func NewDriver(p *Player) *Driver { return &Driver{p: p} }

var (
	soundOnce sync.Once
	sound     *Driver
)

func Sound() *Driver {
	soundOnce.Do(func() { sound = NewDriver(Get()) })
	return sound
}

type Background interface {
	Stand(down bool)
}

func (d *Driver) Yields(b Background) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.bg = b
}

func (d *Driver) Claim(name string, play func(ctx context.Context, p *Player) error) *Claim {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Claim{name: name, cancel: cancel, done: make(chan struct{})}

	d.mu.Lock()
	previous := d.now
	d.now = c
	d.mu.Unlock()

	d.settle()
	previous.preempt(d.p)

	go func() {
		defer close(c.done)
		defer d.release(c)

		if err := play(ctx, d.p); err != nil {
			c.fail(err)
			return
		}
		c.mark(d.await(ctx))
	}()
	return c
}

func (d *Driver) release(c *Claim) {
	d.mu.Lock()
	if d.now != c {
		d.mu.Unlock()
		return
	}
	d.now = nil
	d.mu.Unlock()

	d.settle()
}

func (d *Driver) settle() {
	d.mu.Lock()
	want := d.now != nil
	if want == d.yielded {
		d.mu.Unlock()
		return
	}
	d.yielded = want
	bg := d.bg
	d.mu.Unlock()

	if bg == nil {
		return
	}
	bg.Stand(want)
}

func (d *Driver) Interject(play func(p *Player)) { play(d.p) }

func (d *Driver) Silence() {
	d.mu.Lock()
	c := d.now
	d.now = nil
	d.mu.Unlock()

	c.stop(d.p)
	d.p.Drain()
	d.settle()
}

func (d *Driver) Busy() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.now != nil && !d.now.Finished()
}

const HardwareTail = 150 * time.Millisecond

const dry = 400 * time.Millisecond

func (d *Driver) await(ctx context.Context) time.Time {
	const tick = 50 * time.Millisecond

	var empty time.Time
	for {
		select {
		case <-ctx.Done():
			return time.Now()
		case <-time.After(tick):
		}

		if d.p.Queued() > 0 {
			empty = time.Time{}
			continue
		}
		if empty.IsZero() {
			empty = time.Now()
		}
		if time.Since(empty) >= dry {
			time.Sleep(HardwareTail)
			return empty
		}
	}
}

type Claim struct {
	name   string
	cancel context.CancelFunc
	done   chan struct{}

	mu       sync.Mutex
	err      error
	stopped  bool
	started  time.Time
	finished time.Time
}

func (c *Claim) Started() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started.IsZero() {
		c.started = time.Now()
	}
}

func (c *Claim) Playing() time.Time {
	if c == nil {
		return time.Time{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.started
}

func (c *Claim) Quiet() time.Time {
	if c == nil {
		return time.Time{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.finished
}

func (c *Claim) Done() <-chan struct{} {
	if c == nil {
		closed := make(chan struct{})
		close(closed)
		return closed
	}
	return c.done
}

func (c *Claim) Stopped() bool {
	if c == nil {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stopped
}

func (c *Claim) Err() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *Claim) Finished() bool {
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}

func (c *Claim) preempt(p *Player) {
	if c == nil || c.Finished() {
		return
	}
	c.stop(p)
}

func (c *Claim) stop(p *Player) {
	if c == nil {
		return
	}

	c.mu.Lock()
	c.stopped = true
	c.mu.Unlock()

	c.cancel()
	p.Drain()

	select {
	case <-c.done:
	case <-time.After(time.Second):
		slog.Warn("sound would not stop", "claim", c.name)
	}

	p.Drain()
}

func (c *Claim) mark(quiet time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.finished = quiet
}

func (c *Claim) fail(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
}
