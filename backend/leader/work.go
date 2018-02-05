package leader

import (
	"context"
	"sync/atomic"
)

var (
	workItemSerial int64
)

// work is a single unit of work, performed by f.
type work struct {
	id     int64
	f      func(context.Context) error
	result chan error
	err    error
}

func newWork(f func(context.Context) error) *work {
	return &work{
		id:     newWorkID(),
		f:      f,
		result: make(chan error, 1),
	}
}

func newWorkID() int64 {
	return atomic.AddInt64(&workItemSerial, 1)
}

func (w *work) Err() error {
	err, ok := <-w.result
	if ok {
		w.err = err
		close(w.result)
	}
	return w.err
}
