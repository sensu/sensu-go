package crock

import (
	"sync/atomic"
	"time"

	timeproxy "github.com/echlebek/timeproxy"
)

// NewTimer creates a new crock timer. It works like time.NewTimer except the
// timer only sends
func (t *Time) NewTimer(d Duration) *timeproxy.Timer {
	timer := newTimer(t, d)
	return &timeproxy.Timer{
		C:         timer.C,
		StopFunc:  timer.Stop,
		ResetFunc: timer.Reset,
	}
}

type timer struct {
	running     int64
	time        *Time
	C           chan time.Time
	duration    time.Duration
	fireAt      time.Time
	eventFuncID uint64
	eventFunc   func()
}

func (t *timer) Stop() bool {
	swapped := atomic.CompareAndSwapInt64(&t.running, 1, 0)
	if !swapped {
		return false
	}
	return t.time.cancelEvent(t.fireAt, t.eventFuncID)
}

func newTimer(t *Time, d time.Duration) *timer {
	at := t.Now().Add(d)
	timer := &timer{
		time:     t,
		C:        make(chan time.Time, 1),
		duration: d,
		fireAt:   at,
		running:  1,
	}
	timer.eventFunc = func() {
		timer.C <- t.Now()
		atomic.StoreInt64(&timer.running, 0)
	}
	timer.eventFuncID = t.event(at, timer.eventFunc)
	return timer
}

func newAfterTimer(t *Time, d time.Duration, f func()) *timer {
	at := t.Now().Add(d)
	timer := &timer{
		time:      t,
		C:         make(chan time.Time, 1),
		duration:  d,
		fireAt:    at,
		eventFunc: f,
	}
	timer.eventFuncID = t.event(at, timer.eventFunc)
	timer.running = 1

	return timer
}

func (t *timer) Reset(d time.Duration) bool {
	active := t.Stop()
	at := t.time.Now().Add(d)
	t.time.event(at, t.eventFunc)
	t.fireAt = at

	atomic.StoreInt64(&t.running, 1)

	return active
}

func (t *Time) AfterFunc(d Duration, f func()) *timeproxy.Timer {
	timer := newAfterTimer(t, d, f)
	return &timeproxy.Timer{
		C:         timer.C,
		StopFunc:  timer.Stop,
		ResetFunc: timer.Reset,
	}
}
