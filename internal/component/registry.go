package component

import (
	"sync"

	sharedcomponent "github.com/ygelfand/libcountertop/pkg/runtime/component"

	"github.com/ygelfand/echolocal/internal/board"
	"github.com/ygelfand/echolocal/internal/config"
)

type Registry struct {
	*sharedcomponent.Registry[config.Config]
	mu     sync.Mutex
	detect sync.Once
	board  board.Board
	told   bool
}

func New() *Registry {
	r := &Registry{}
	r.Registry = sharedcomponent.New[config.Config](
		func(needs uint64) bool { return needs & ^uint64(r.Board().Caps) == 0 },
	)
	return r
}
func Needs(c board.Cap) sharedcomponent.Option { return sharedcomponent.Needs(uint64(c)) }

func (r *Registry) Use(
	b board.Board,
) {
	r.mu.Lock()
	r.board, r.told = b, true
	r.mu.Unlock()
}
func (r *Registry) Board() board.Board {
	r.detect.Do(func() {
		r.mu.Lock()
		told := r.told
		r.mu.Unlock()
		if !told {
			b := board.Detect()
			r.mu.Lock()
			if !r.told {
				r.board, r.told = b, true
			}
			r.mu.Unlock()
		}
	})
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.board
}
func Board() board.Board { return shared.Board() }

var shared = New()

func Default() *Registry { return shared }

func Register[T sharedcomponent.Component](
	p sharedcomponent.Phase,
	make func() T,
	opts ...sharedcomponent.Option,
) {
	sharedcomponent.Register(shared.Registry, p, make, opts...)
}
