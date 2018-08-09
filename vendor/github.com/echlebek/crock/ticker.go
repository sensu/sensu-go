package crock

import (
	"runtime"
	"sync/atomic"
	"time"

	timeproxy "github.com/echlebek/timeproxy"
)

// NewTicker creates a new crock ticker. It works like time.Ticker, except that
// it only ticks when time is progressing.
func (t *Time) NewTicker(d time.Duration) *timeproxy.Ticker {
	ticker := newTicker(t, d)
	// start guarantees that ticker will be reachable until stop is called
	ticker.start()
	// Make sure we don't keep creating tick events after the ticker has gone
	// out of scope.
	stopper := ticker.Stop
	finalizer := func(interface{}) {
		stopper()
	}
	runtime.SetFinalizer(ticker, finalizer)
	return &timeproxy.Ticker{
		C:        ticker.C,
		StopFunc: ticker.Stop,
	}
}

func newTicker(t *Time, d time.Duration) *ticker {
	return &ticker{
		time:     t,
		duration: d,
		C:        make(chan time.Time, 1),
	}
}

func (t *ticker) start() {
	t.running = 1
	now := t.time.Now()
	f := new(func())
	// use a pointer to have this closure add itself to time events.
	// the magic of indirection!
	*f = func() {
		now := t.time.Now()
		if atomic.LoadInt64(&t.running) == 1 {
			t.time.event(now.Add(t.duration), *f)
		}
		t.C <- now
	}
	t.time.event(now.Add(t.duration), *f)
}

// Ticker stops the ticker. No new events will be sent on ticker.C.
func (t *ticker) Stop() {
	atomic.StoreInt64(&t.running, 0)
}

// Ticker is like time.Ticker, but will tick only when time is progressing
type ticker struct {
	running  int64
	time     *Time
	duration time.Duration
	C        chan time.Time
}
