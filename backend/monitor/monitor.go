package monitor

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// Monitor is a managed timer that is reset whenever the monitor observes a
// Keepalive or Check Result Ttl event via the Update() function. Once the timer
// has been stopped, it cannot be started or used again.
type Monitor struct {
	Entity         *types.Entity
	Timeout        time.Duration
	FailureHandler MonitorFailureHandler
	UpdateHandler  MonitorUpdateHandler
	Deregisterer   Deregisterer
	EventCreator   EventCreator
	MessageBus     messaging.MessageBus
	Store          store.Store

	reset   chan interface{}
	timer   *time.Timer
	stopped int32
	failing int32
}

/* this stuff should happen in the start function
func foo() {
	creator := &KeepaliveEventCreator(entity)
	handler := &keepaliveUpdateHandler{}

	monitor := &Monitor{
		Timeout:        60 * time.Second,
		FailureHandler: handler,
		UpdateHandler:  handler,
	}
}
*/

type MonitorFailureHandler interface {
	HandleFailure(t time.T) error
}

type MonitorUpdateHandler interface {
	HandleUpdate(e *types.Event) error
}

// Start initializes the monitor and starts its monitoring goroutine.
// If the monitor has been previously stopped, this method has no
// effect.
func (monitorPtr *Monitor) Start() {
	if monitorPtr.IsStopped() {
		return
	}

	// refactor this to take a timeout (keepalive or check ttl)
	timerDuration := monitorPtr.Timeout * time.Second
	monitorPtr.timer = time.NewTimer(timerDuration)
	monitorPtr.reset = make(chan interface{})
	go func() {
		timer := monitorPtr.timer

		var (
			err     error
			timeout int64
		)

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

				atomic.CompareAndSwapInt32(&monitorPtr.failing, 0, 1)
			}

			if monitorPtr.IsStopped() {
				return
			}

			timer.Reset(timerDuration)
		}
	}()
}

// HandleUpdate causes the Monitor to observe the event. If the monitor has
// been stopped, this method has no effect.
func (monitorPtr *Monitor) HandleUpdate(event *types.Event) error {
	// once the monitor is stopped, we can't continue, because the
	// reset channel will be closed.
	if monitorPtr.IsStopped() {
		return nil
	}

	monitorPtr.UpdateHandler.HandleUpdate(event)
	monitorPtr.reset <- struct{}{}
	return nil
}

// HandleFailure
func (monitorPtr *Monitor) HandleFailure(event *types.Event) error {
	return monitorPtr.FailureHandler.HandleFailure(event)
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
