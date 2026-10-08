package led

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/hardware/indicate"
	"github.com/ygelfand/echolocal/internal/service"
)

type Priority = indicate.Priority

const (
	PriorityBase    = indicate.PriorityBase
	PriorityRoom    = indicate.PriorityRoom
	PriorityMute    = indicate.PriorityMute
	PriorityTimer   = indicate.PriorityTimer
	PriorityBusy    = indicate.PriorityBusy
	PriorityTurn    = indicate.PriorityTurn
	PriorityAlarm   = indicate.PriorityAlarm
	PriorityNotice  = indicate.PriorityNotice
	PriorityTrouble = indicate.PriorityTrouble
	PriorityBoot    = indicate.PriorityBoot
)

type Content struct {
	Frame []Color

	Effect  string
	Base    Color
	Reverse bool

	Room Room

	Animate func(ctx context.Context, r *Ring) error
}

func (c Content) empty() bool {
	return c.Animate == nil && c.Effect == "" && len(c.Frame) == 0
}

type Driver struct {
	ring *Ring

	busy *Busy

	mu     sync.Mutex
	claims []*Claim
	seq    uint64

	// The hardware keeps the last frame with nothing driving it.
	hold Color

	changed chan struct{}
}

func NewDriver(r *Ring) *Driver {
	d := &Driver{ring: r, changed: make(chan struct{}, 1)}

	d.busy = &Busy{claim: d.Claim(PriorityBusy)}
	return d
}

var (
	once   sync.Once
	shared *Driver
)

func init() {
	component.Register(component.Hardware, Get, component.Order(5),
		component.Needs(board.Ring),
		component.Supervise(service.Restart(time.Second, 30*time.Second)))
}

func Get() *Driver {
	once.Do(func() {
		if !component.Board().Has(board.Ring) {
			shared = NewDriver(NewAbsent())
			return
		}
		shared = NewDriver(New())
	})
	return shared
}

// led_current is an attenuation index: 0 is full current, 3 a quarter and the probe default.
func (d *Driver) Start(context.Context) error {
	d.ring.Forget()

	if err := d.ring.SetBootAnimation(false); err != nil {
		slog.Error("disabling driver boot animation failed", "err", err)
	}
	if err := d.ring.SetCurrent(0); err != nil {
		slog.Error("setting led_current failed", "err", err)
	}
	return nil
}

func (d *Driver) HoldOnStop(c Color) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.hold = c
}

func (d *Driver) Close() error {
	d.mu.Lock()
	hold := d.hold
	d.mu.Unlock()

	return d.ring.SetSegments(Solid(hold))
}

func (d *Driver) Busy() *Busy { return d.busy }

func (d *Driver) Present() bool { return d.ring.Present() }

func (c *Claim) Present() bool { return c.d.Present() }

func (d *Driver) Name() string { return "ring" }

func (d *Driver) Claim(p Priority) *Claim {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.seq++
	c := &Claim{d: d, priority: p, seq: d.seq}
	d.claims = append(d.claims, c)

	return c
}

func (d *Driver) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		top, rev := d.top()
		d.render(ctx, top, rev)

		if top != nil && top.done() {
			top.Release()
		}
	}
}

func (d *Driver) render(ctx context.Context, top *Claim, rev uint64) {
	rctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if top != nil {
		if until, ok := top.deadline(); ok {
			rctx, cancel = context.WithDeadline(ctx, until)
			defer cancel()
		}
	}

	under := d.composited(top)
	underRev := under.revision()

	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-d.changed:
				if now, nowRev := d.top(); now != top || nowRev != rev {
					cancel()
					return
				}
				if under.revision() != underRev {
					cancel()
					return
				}
			case <-rctx.Done():
				return
			case <-stop:
				return
			}
		}
	}()

	if top == nil {
		if err := d.ring.Off(); err != nil {
			slog.Error("ring blank failed", "err", err)
		}
		<-rctx.Done()
		return
	}
	d.show(rctx, top.get(), d.beneath(under))
}

func (d *Driver) show(ctx context.Context, c Content, under Frame) {
	switch {
	case c.Animate != nil:
		if err := c.Animate(ctx, d.ring); err != nil && ctx.Err() == nil {
			slog.Error("ring animation failed", "err", err)
		}
		<-ctx.Done()

	case c.Effect != "":
		var err error
		if c.Reverse {
			err = RunEffectReversed(ctx, d.ring, c.Effect, c.Base, c.Room, under)
		} else {
			err = RunEffect(ctx, d.ring, c.Effect, c.Base, c.Room, under)
		}
		if err != nil && ctx.Err() == nil {
			slog.Error("ring effect failed", "effect", c.Effect, "err", err)
			<-ctx.Done()
		}

	default:
		if err := d.ring.SetSegments(c.Frame); err != nil {
			slog.Error("ring write failed", "err", err)
		}
		<-ctx.Done()
	}
}

func (d *Driver) stack() []*Claim {
	d.mu.Lock()
	defer d.mu.Unlock()

	live := make([]*Claim, 0, len(d.claims))
	for _, c := range d.claims {
		if c.live() {
			live = append(live, c)
		}
	}

	sort.SliceStable(live, func(i, j int) bool {
		if live[i].priority != live[j].priority {
			return live[i].priority > live[j].priority
		}
		return live[i].seq > live[j].seq
	})
	return live
}

func (d *Driver) top() (*Claim, uint64) {
	live := d.stack()
	if len(live) == 0 {
		return nil, 0
	}

	best := live[0]
	best.mu.Lock()
	defer best.mu.Unlock()
	return best, best.rev
}

func (d *Driver) composited(top *Claim) *Claim {
	if top == nil || top.get().Room.Level == nil {
		return nil
	}

	live := d.stack()
	if len(live) < 2 || live[0] != top {
		return nil
	}
	return live[1]
}

func (d *Driver) beneath(c *Claim) Frame {
	if c == nil {
		return nil
	}

	under := c.get()
	switch {
	case under.Animate != nil:
		return nil

	case under.Effect != "":
		frame, err := effect(under.Effect, under.Base, under.Room)
		if err != nil {
			return nil
		}
		return frame

	case len(under.Frame) == Segments:
		still := under.Frame
		return func(time.Duration) []Color { return still }
	}
	return nil
}

func (d *Driver) wake() {
	select {
	case d.changed <- struct{}{}:
	default:
	}
}

func (d *Driver) drop(c *Claim) {
	d.mu.Lock()
	for i, held := range d.claims {
		if held == c {
			d.claims = append(d.claims[:i], d.claims[i+1:]...)
			break
		}
	}
	d.mu.Unlock()
	d.wake()
}

type Claim struct {
	d        *Driver
	priority Priority
	seq      uint64

	mu       sync.Mutex
	content  Content
	expires  time.Time
	released bool

	rev uint64
}

func (c *Claim) Show(content Content) { c.set(content, time.Time{}) }

func (c *Claim) ShowFor(effect string, base Color, d time.Duration) {
	c.ShowContentFor(Content{Effect: effect, Base: base}, d)
}

func (c *Claim) ShowContentFor(content Content, d time.Duration) {
	c.set(content, time.Now().Add(d))
}

func (c *Claim) Paint(frame []Color) { c.Show(Content{Frame: frame}) }

func (c *Claim) PaintFor(frame []Color, d time.Duration) {
	c.ShowContentFor(Content{Frame: frame}, d)
}

func (c *Claim) Play(effect string, base Color) { c.Show(Content{Effect: effect, Base: base}) }

func (c *Claim) React(effect string, base Color, room Room) {
	c.Show(Content{Effect: effect, Base: base, Room: room})
}

func (c *Claim) PlayReversed(effect string, base Color) {
	c.Show(Content{Effect: effect, Base: base, Reverse: true})
}

func (c *Claim) Clear() { c.set(Content{}, time.Time{}) }

func (c *Claim) set(content Content, expires time.Time) {
	c.mu.Lock()
	if c.released {
		c.mu.Unlock()
		return
	}
	c.content, c.expires = content, expires
	c.rev++
	c.mu.Unlock()

	c.d.wake()
}

func (c *Claim) Release() {
	c.mu.Lock()
	if c.released {
		c.mu.Unlock()
		return
	}
	c.released = true
	c.content = Content{}
	c.mu.Unlock()

	c.d.drop(c)
}

func (c *Claim) live() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.released && !c.content.empty()
}

func (c *Claim) get() Content {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.content
}

func (c *Claim) revision() uint64 {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.rev
}

func (c *Claim) deadline() (time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.expires, !c.expires.IsZero()
}

func (c *Claim) done() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.expires.IsZero() && !time.Now().Before(c.expires)
}

type surface struct{ d *Driver }

var _ indicate.Surface = surface{}
var _ indicate.Claim = (*Claim)(nil)

func (s surface) Claim(p indicate.Priority) indicate.Claim { return s.d.Claim(p) }
func (s surface) HoldOnStop(c indicate.Color)              { s.d.HoldOnStop(c) }
func (s surface) Present() bool                            { return s.d.Present() }

func init() {
	indicate.Register(func() indicate.Surface { return surface{Get()} })
}
