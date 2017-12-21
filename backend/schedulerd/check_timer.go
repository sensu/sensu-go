package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"time"
)

// A CheckTimer handles starting and stopping timers for a given check
type CheckTimer interface {
	NewIntervalTimer(name string, interval uint) *IntervalTimer
}

// A IntervalTimer handles starting a stopping timers for a given check
type IntervalTimer struct {
	interval time.Duration
	splay    uint64
	timer    *time.Timer
}

// NewIntervalTimer establishes new check timer given a name & an initial interval
func NewIntervalTimer(name string, interval uint) *IntervalTimer {
	// Calculate a check execution splay to ensure
	// execution is consistent between process restarts.
	sum := md5.Sum([]byte(name))
	splay := binary.LittleEndian.Uint64(sum[:])

	timer := &IntervalTimer{splay: splay}
	timer.SetInterval(interval)
	return timer
}

// C channel emits events when timer's duration has reaached 0
func (timerPtr *IntervalTimer) C() <-chan time.Time {
	return timerPtr.timer.C
}

// SetInterval updates the interval in which timers are set
func (timerPtr *IntervalTimer) SetInterval(i uint) {
	timerPtr.interval = time.Duration(time.Second * time.Duration(i))
}

// Start sets up a new timer
func (timerPtr *IntervalTimer) Start() {
	initOffset := timerPtr.calcInitialOffset()
	timerPtr.timer = time.NewTimer(initOffset)
}

// Next reset's timer using interval
func (timerPtr *IntervalTimer) Next() {
	timerPtr.timer.Reset(timerPtr.interval)
}

// Stop ends the timer
func (timerPtr *IntervalTimer) Stop() bool {
	return timerPtr.timer.Stop()
}

// Calculate the first execution time using splay & interval
func (timerPtr *IntervalTimer) calcInitialOffset() time.Duration {
	now := uint64(time.Now().UnixNano())
	offset := (timerPtr.splay - now) % uint64(timerPtr.interval)
	return time.Duration(offset) / time.Nanosecond
}
