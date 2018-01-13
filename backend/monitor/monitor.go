package monitor

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/types"
)

// Interface ...
type Interface interface {
	Stop()
	IsStopped() bool
	HandleUpdate(e *types.Event) error
	HandleFailure(e *types.Entity) error
}

// Monitor is a managed timer that is reset whenever the monitor observes a
// Keepalive or Check Result Ttl event via the Update() function. Once the timer
// has been stopped, it cannot be started or used again.
type Monitor struct {
	Entity         *types.Entity
	Timeout        time.Duration
	FailureHandler FailureHandler
	UpdateHandler  UpdateHandler

	resetChan chan time.Duration
	timer     *time.Timer
	stopped   int32
	failing   int32
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
func (m *Monitor) HandleUpdate(event *types.Event) error {
	// once the monitor is stopped, we can't continue, because the
	// reset channel will be closed.
	if m.IsStopped() {
		return nil
	}

	// Update the event. If the monitor is failing, flip status back to zero
	atomic.CompareAndSwapInt32(&m.failing, 1, 0)
	m.reset(m.Timeout * time.Second)
	return m.UpdateHandler.HandleUpdate(event)
}

// HandleFailure passes an event to the failure handler function and runs it.
func (m *Monitor) HandleFailure(entity *types.Entity) error {
	fmt.Printf("Entity: %v\n", entity)
	fmt.Println("In monitor failure handler")
	defer m.Stop()
	atomic.CompareAndSwapInt32(&m.failing, 0, 1)
	return m.FailureHandler.HandleFailure(entity)
}

// start initializes the monitor and starts its monitoring goroutine.
// If the monitor has been previously stopped, this method has no
// effect.
func (m *Monitor) start() {
	if m.IsStopped() {
		return
	}

	timerDuration := m.Timeout * time.Second
	m.timer = time.NewTimer(timerDuration)
	m.resetChan = make(chan time.Duration)
	go func() {
		timer := m.timer

		for {
			// Access to the timer has to be constrained to a single goroutine.
			// Otherwise, we have an unavoidable race between reading from timer.C
			// and calling timer.Reset(), so we signal a clean reset of the
			// timer using the reset channel.
			select {
			case d := <-m.resetChan:
				if d == 0 {
					if !timer.Stop() {
						<-timer.C
					}
					timer.Reset(timerDuration)
				} else {
					timer.Reset(d)
				}

			case <-timer.C:
				m.HandleFailure(m.Entity)
				timer.Reset(timerDuration)
			}

			if m.IsStopped() {
				return
			}

		}
	}()
}

// Stop the Monitor. Once the monitor has been stopped it
// can no longer be used.
func (m *Monitor) Stop() {
	// atomically set stopped so that once Stop is called, all future
	// reads of stopped are true.
	if !atomic.CompareAndSwapInt32(&m.stopped, 0, 1) {
		return
	}

	close(m.resetChan)
}

// IsStopped returns true if the Monitor has been stopped.
func (m *Monitor) IsStopped() bool {
	return atomic.LoadInt32(&m.stopped) > 0
}

// Reset the Monitor's timer to emit an event at a given time.
// Once the Monitor has been stopped, this has no effect.
func (m *Monitor) reset(t time.Duration) {
	if m.IsStopped() {
		return
	}

	m.resetChan <- t
}

// New creates a new monitor from an entity.
func New(e *types.Entity, t time.Duration, updateHandler UpdateHandler, failureHandler FailureHandler) *Monitor {
	monitor := &Monitor{
		Entity:         e,
		Timeout:        t,
		FailureHandler: failureHandler,
		UpdateHandler:  updateHandler,
	}
	monitor.start()
	return monitor
}
