package schedulerd

import (
	"crypto/md5"
	"encoding/binary"
	"time"

	"github.com/robfig/cron"
)

// A CheckTimer handles starting and stopping timers for a given check
type CheckTimer interface {
	// C channel emits events when timer's duration has reached 0
	C() <-chan time.Time
	// SetDuration updates the interval in which timers are set
	SetDuration(string, uint)
	// Start sets up a new timer
	Start()
	// Next reset's timer using interval
	Next()
	// Stop ends the timer
	Stop()
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
	timer.SetDuration("", interval)
	return timer
}

// C channel emits events when timer's duration has reached 0
func (timerPtr *IntervalTimer) C() <-chan time.Time {
	return timerPtr.timer.C
}

// SetDuration updates the interval in which timers are set
func (timerPtr *IntervalTimer) SetDuration(cron string, interval uint) {
	timerPtr.interval = time.Duration(time.Second * time.Duration(interval))
}

// Start sets up a new timer
func (timerPtr *IntervalTimer) Start() {
	initOffset := timerPtr.calcInitialOffset()
	timerPtr.timer = time.NewTimer(initOffset)
}

// Next reset's timer using interval
func (timerPtr *IntervalTimer) Next() {
	if !timerPtr.timer.Reset(timerPtr.interval) {
		<-timerPtr.timer.C
	}
}

// Stop ends the timer
func (timerPtr *IntervalTimer) Stop() {
	if !timerPtr.timer.Stop() {
		<-timerPtr.timer.C
	}
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
func NewCronTimer(name string, cronStr string) *CronTimer {
	diff, err := NextCronTime(time.Now(), cronStr)
	// we shouldn't hit this error because we've already validated the cron string
	// but log and exit cleanly to revert to the interval timer
	if err != nil {
		logger.WithError(err).Error("invalid cron, reverting to interval")
		return nil
	}
	timer := &CronTimer{next: diff}
	return timer
}

// C channel emits events when timer's duration has reached 0
func (timerPtr *CronTimer) C() <-chan time.Time {
	return timerPtr.timer.C
}

// SetDuration updates the interval in which timers are set
func (timerPtr *CronTimer) SetDuration(cronStr string, interval uint) {
	diff, err := NextCronTime(time.Now(), cronStr)
	// we shouldn't hit this error because we've already validated the cron string
	// but log and exit cleanly to revert to the interval timer
	if err != nil {
		logger.WithError(err).Error("invalid cron, reverting to interval")
		return
	}
	timerPtr.next = diff
}

// Start sets up a new timer
func (timerPtr *CronTimer) Start() {
	timerPtr.timer = time.NewTimer(timerPtr.next)
}

// Next reset's timer using interval
func (timerPtr *CronTimer) Next() {
	if !timerPtr.timer.Reset(timerPtr.next) {
		<-timerPtr.timer.C
	}
}

// Stop ends the timer
func (timerPtr *CronTimer) Stop() {
	if !timerPtr.timer.Stop() {
		<-timerPtr.timer.C
	}
}

// NextCronTime calculates how much time is between the current time and the
// time indidcated by the cron string
func NextCronTime(now time.Time, cronStr string) (time.Duration, error) {
	schedule, err := cron.ParseStandard(cronStr)
	if err != nil {
		return 0, err
	}
	nextTime := schedule.Next(now)
	diff := nextTime.Sub(now)

	return diff, nil
}
