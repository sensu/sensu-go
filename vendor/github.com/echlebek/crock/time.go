package crock

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	// DefaultResolution is the resolution at which crock time updates.
	DefaultResolution = time.Second

	// DefaultMultiplier is the relative speed at which crock time progresses.
	DefaultMultiplier = 1.0
)

var (
	funcIDCounter uint64
)

// Time is a crock implementation of time. New Times are halted by default -
// they do not advance. Time proceeds according to the resolution and
// multiplier settings. By default, the Resolution is DefaultResolution
// and Multiplier is DefaultMultiplier.
//
// Don't change the Resolution and Multiplier fields when time is in motion.
type Time struct {
	running int64
	now     time.Time
	mu      sync.Mutex
	done    chan struct{}
	events  map[int64][]idFunc

	// Resolution is the frequency events will be processed at once crock time
	// is started.
	Resolution time.Duration

	// Multiplier controls the relative speed of crock time. A multiplier of
	// 1.0 means crock time will proceed at nearly the same speed as real time.
	Multiplier float64
}

// NewTime creates a new time, which is now. Resolution and Multiplier are
// set to their defaults.
func NewTime(now time.Time) *Time {
	return &Time{
		now:        now,
		Resolution: DefaultResolution,
		Multiplier: DefaultMultiplier,
		events:     make(map[int64][]idFunc),
	}
}

// Now returns t's current time.
func (t *Time) Now() time.Time {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.now
}

// Sleep sleeps for real duration d / t.Multiplier.
func (t *Time) Sleep(d time.Duration) {
	time.Sleep(t.duration(d))
}

// Start causes crock time to progress at the rate determined by t.Multiplier.
// For example, if t.Multiplier is 0.5, crock time will progress twice as slowly
// as real time.
//
// Don't call Start concurrently with other methods.
func (t *Time) Start() {
	if atomic.CompareAndSwapInt64(&t.running, 0, 1) {
		t.done = make(chan struct{})
		go t.loop(t.done)
	}
}

// Stop stops crock time.
//
// Don't call Stop concurrently with other methods.
func (t *Time) Stop() {
	if atomic.CompareAndSwapInt64(&t.running, 1, 0) {
		close(t.done)
	}
}

// Set sets crock time to a particular time. Can be invoked whether or not time
// is currently progressing. If there are timer or ticker events that would
// have occurred before the set crock time, they will fire.
func (t *Time) Set(to time.Time) {
	t.mu.Lock()
	t.now = to
	t.mu.Unlock()
	t.processEvents()
}

// time event loop
func (t *Time) loop(done chan struct{}) {
	ticker := time.NewTicker(t.Resolution)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			t.Set(t.Now().Add(time.Duration(float64(t.Resolution) * t.Multiplier)))
			go t.processEvents()
		case <-done:
			return
		}
	}
}

// duration = d / t.Multiplier
func (t *Time) duration(d time.Duration) time.Duration {
	return time.Duration(float64(d) / t.Multiplier)
}

type idFunc struct {
	id uint64
	f  func()
}

func (f idFunc) Call() {
	f.f()
}

// event registers an event to be executed at a time.
func (t *Time) event(at time.Time, do func()) uint64 {
	t.mu.Lock()
	defer t.mu.Unlock()

	fn := idFunc{
		id: atomic.AddUint64(&funcIDCounter, 1),
		f:  do,
	}
	t.events[at.UnixNano()] = append(t.events[at.UnixNano()], fn)

	return fn.id
}

func (t *Time) cancelEvent(at time.Time, id uint64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	events := t.events[at.UnixNano()]
	newEvents := []idFunc{}
	for _, e := range events {
		if e.id == id {
			continue
		}
		newEvents = append(newEvents, e)
	}
	t.events[at.UnixNano()] = newEvents

	return len(newEvents) < len(events)
}

// processEvents loops over all of the events that are registered,
// executes them, and then deletes them.
func (t *Time) processEvents() {
	now := t.Now().UnixNano()
	t.mu.Lock()
	defer t.mu.Unlock()
	for nano, funcs := range t.events {
		if nano <= now {
			for i, fn := range funcs {
				t.events[nano] = append(t.events[nano][:i], t.events[nano][i+1:]...)
				go fn.Call()
			}
		}
	}
}

// After works like time.After, except it only sends on the channel it returns
// if crock time progresses enough.
func (t *Time) After(d time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)

	now := t.Now()
	then := now.Add(d)

	t.event(then, func() { ch <- then; close(ch) })

	return ch
}

// Tick works like time.Tick, except it only ticks if crock time progresses
// enough.
func (t *Time) Tick(d time.Duration) <-chan time.Time {
	return t.NewTicker(d).C
}

// Since works like time.Since, performing t.Now().Sub(s)
func (t *Time) Since(s time.Time) time.Duration {
	return t.Now().Sub(s)
}

// Until works like time.Until, performing s.Sub(t.Now())
func (t *Time) Until(s time.Time) time.Duration {
	return s.Sub(t.Now())
}
