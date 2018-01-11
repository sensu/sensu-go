package monitor

import (
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/types"
)

// Monitor is a managed timer that is reset whenever the monitor observes a
// Keepalive or Check Result Ttl event via the Update() function. Once the timer
// has been stopped, it cannot be started or used again.
type Monitor struct {
	Entity         *types.Entity
	Timeout        time.Duration
	FailureHandler FailureHandler
	UpdateHandler  UpdateHandler

	reset   chan interface{}
	timer   *time.Timer
	stopped int32
	failing int32
}

// UpdateHandler provides a HandleUpdate function.
type UpdateHandler interface {
	HandleUpdate(e *types.Event) error
}

// FailureHandler provides a HandleFailure function.
type FailureHandler interface {
	HandleFailure(e *types.Entity) error
}

// HandleUpdate causes the Monitor to observe the event. If the monitor has
// been stopped, this method has no effect.
func (monitorPtr *Monitor) HandleUpdate(event *types.Event) error {
	// once the monitor is stopped, we can't continue, because the
	// reset channel will be closed.
	if monitorPtr.IsStopped() {
		return nil
	}

	if atomic.CompareAndSwapInt32(&monitorPtr.failing, 1, 0) {
		if err := monitorPtr.UpdateHandler.HandleUpdate(event); err != nil {
			return err
		}
	}
	monitorPtr.reset <- struct{}{}
	return nil
}

// HandleFailure passes an event to the failure handler function and runs it.
func (monitorPtr *Monitor) HandleFailure(entity *types.Entity) error {
	return monitorPtr.FailureHandler.HandleFailure(entity)
}

// Start initializes the monitor and starts its monitoring goroutine.
// If the monitor has been previously stopped, this method has no
// effect.
func (monitorPtr *Monitor) Start() {
	if monitorPtr.IsStopped() {
		return
	}

	timerDuration := monitorPtr.Timeout * time.Second
	monitorPtr.timer = time.NewTimer(timerDuration)
	monitorPtr.reset = make(chan interface{})
	go func() {
		timer := monitorPtr.timer

		for {
			// Access to the timer has to be constrained to a single goroutine.
			// Otherwise, we have an unavoidable race between reading from timer.C
			// and calling timer.Reset(), so we signal a clean reset of the
			// timer using the reset channel.
			select {
			case <-monitorPtr.reset:
				if !timer.Stop() {
					<-timer.C
				}

			case <-timer.C:
				// check the event deregistration; if so, delete and return
				// otherwise - emit an event and then swap
				monitorPtr.FailureHandler.HandleFailure(monitorPtr.Entity)

				atomic.CompareAndSwapInt32(&monitorPtr.failing, 0, 1)
			}

			if monitorPtr.IsStopped() {
				return
			}

			timer.Reset(timerDuration)
		}
	}()
}

// Stop the Monitor. Once the monitor has been stopped it
// can no longer be used.
func (monitorPtr *Monitor) Stop() {
	// atomically set stopped so that once Stop is called, all future
	// reads of stopped are true.
	if !atomic.CompareAndSwapInt32(&monitorPtr.stopped, 0, 1) {
		return
	}

	close(monitorPtr.reset)
}

// IsStopped returns true if the Monitor has been stopped.
func (monitorPtr *Monitor) IsStopped() bool {
	return atomic.LoadInt32(&monitorPtr.stopped) > 0
}

// Reset the Monitor's timer to emit an event at a given time.
// Once the Monitor has been stopped, this has no effect.
func (monitorPtr *Monitor) Reset(t int64) {
	if monitorPtr.IsStopped() {
		return
	}

	if monitorPtr.timer == nil {
		monitorPtr.Start()
	}

	d := time.Duration(t - time.Now().Unix())
	if d < 0 {
		d = 0
	}

	monitorPtr.timer.Reset(d)
}
