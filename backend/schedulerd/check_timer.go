package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"time"

	"github.com/robfig/cron"
)

// A CheckTimer handles starting and stopping timers for a given check
type CheckTimer interface {
	C() <-chan time.Time
	SetDuration(uint)
	Start()
	Next()
	Stop() bool
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
	timer.SetDuration(interval)
	return timer
}

// C channel emits events when timer's duration has reaached 0
func (timerPtr *IntervalTimer) C() <-chan time.Time {
	return timerPtr.timer.C
}

// SetDuration updates the interval in which timers are set
func (timerPtr *IntervalTimer) SetDuration(i uint) {
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

// A CronTimer handles starting and stopping timers for a given check
type CronTimer struct {
	next  time.Duration
	timer *time.Timer
}

// NewCronTimer establishes new check timer given a name & an initial interval
func NewCronTimer(name string, schedule cron.Schedule) *CronTimer {
	nowTime := time.Now()
	nextTime := schedule.Next(nowTime)
	diff := nextTime.Sub(nowTime)
	timer := &CronTimer{next: diff}
	return timer
}

// C channel emits events when timer's duration has reaached 0
func (timerPtr *CronTimer) C() <-chan time.Time {
	return timerPtr.timer.C
}

// SetDuration updates the interval in which timers are set
func (timerPtr *CronTimer) SetDuration(i uint) {
	timerPtr.next = time.Duration(time.Second * time.Duration(i))
}

// Start sets up a new timer
func (timerPtr *CronTimer) Start() {
	timerPtr.timer = time.NewTimer(0)
}

// Next reset's timer using interval
func (timerPtr *CronTimer) Next() {
	timerPtr.timer.Reset(timerPtr.next)
}

// Stop ends the timer
func (timerPtr *CronTimer) Stop() bool {
	return timerPtr.timer.Stop()
}
