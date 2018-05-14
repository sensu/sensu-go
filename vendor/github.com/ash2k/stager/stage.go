package stager

import (
	"context"

	"github.com/ash2k/stager/wait"
)

type Stage interface {
	// Start starts f in a new goroutine attached to the Stage.
	Start(func())
	// StartWithChannel starts f in a new goroutine attached to the Stage.
	// Stage context's Done() channel is passed to f as an argument. f should stop when it is available.
	StartWithChannel(func(stopCh <-chan struct{}))
	// StartWithContext starts f in a new goroutine attached to the Stage.
	// Stage context is passed to f as an argument. f should stop when context's Done() channel is available.
	StartWithContext(func(context.Context))
}

type stage struct {
	ctx    context.Context
	cancel context.CancelFunc
	group  wait.Group
}

func (s *stage) Start(f func()) {
	s.group.Start(f)
}

func (s *stage) StartWithChannel(f func(stopCh <-chan struct{})) {
	s.group.StartWithChannel(s.ctx.Done(), f)
}

func (s *stage) StartWithContext(f func(context.Context)) {
	s.group.StartWithContext(s.ctx, f)
}
