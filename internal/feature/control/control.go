package control

import (
	"context"
	"sync"
	"time"

	harness "github.com/ygelfand/libcountertop/pkg/runtime/control"

	"github.com/ygelfand/echolocal/internal/component"
	"github.com/ygelfand/echolocal/internal/service"
)

const Socket = "@echod"

type Control struct {
	inputOnce sync.Once
	input     *harness.Input
	gate      sync.Mutex
	runtime   *harness.Server
	addr      string
}

var (
	once   sync.Once
	shared *Control
)

func init() {
	component.Register(
		component.Device,
		Get,
		component.Order(90),
		component.Supervise(service.Restart(time.Second, time.Minute)),
	)
}
func Get() *Control             { once.Do(func() { shared = build() }); return shared }
func build() *Control           { return &Control{addr: Socket} }
func (c *Control) Name() string { return "control" }
func (c *Control) Run(ctx context.Context) error {
	if err := c.up(); err != nil {
		return err
	}
	<-ctx.Done()
	c.down()
	return nil
}

func (c *Control) server() *harness.Server {
	c.gate.Lock()
	defer c.gate.Unlock()
	if c.runtime == nil {
		c.runtime = harness.NewServer(
			c.addr,
			func(ctx context.Context, args []string) (string, error) { return harness.Execute(ctx, c.tree, args) },
		)
	}
	return c.runtime
}
func (c *Control) up() error { return c.server().Start() }
func (c *Control) down() {
	c.gate.Lock()
	server := c.runtime
	c.gate.Unlock()
	if server != nil {
		server.Close()
	}
}
