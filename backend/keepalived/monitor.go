package keepalived

import (
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// A Monitor observes events and takes actions based on them.
type Monitor interface {
	Update(event *types.Event) error
	Start()
	Stop()
	IsStopped() bool
}

// KeepaliveMonitor is a managed timer that is reset whenever the monitor
// observes a Keepalive event via the Update() function.
type KeepaliveMonitor struct {
	Entity       *types.Entity
	Deregisterer Deregisterer
	EventCreator EventCreator
	Store        store.Store

	reset   chan interface{}
	timer   *time.Timer
	stopped int32
}

// Start initializes the monitor and starts its monitoring goroutine.
func (monitorPtr *KeepaliveMonitor) Start() {
	timerDuration := time.Duration(monitorPtr.Entity.KeepaliveTimeout) * time.Second
	monitorPtr.timer = time.NewTimer(timerDuration)
	monitorPtr.reset = make(chan interface{})

	running := make(chan interface{})
	go func() {
		close(running)
		timer := monitorPtr.timer
		for {
			select {
			case <-monitorPtr.reset:
				if !timer.Stop() {
					<-timer.C
				}
				if monitorPtr.IsStopped() {
					return
				}

			case <-timer.C:
				// timed out keepalive
				if monitorPtr.Entity.Deregister {
					if err := monitorPtr.Deregisterer.Deregister(monitorPtr.Entity); err != nil {
						logger.WithError(err).Error("error deregistering entity")
					}
					monitorPtr.Stop()
					return
				}
				if err := monitorPtr.EventCreator.Warn(monitorPtr.Entity); err != nil {
					logger.WithError(err).Error("error sending keepalive event")
				}
			}
			timer.Reset(timerDuration)
		}
	}()
	<-running
}

// Update causes the KeepaliveMonitor to observe the event.
func (monitorPtr *KeepaliveMonitor) Update(event *types.Event) error {
	monitorPtr.reset <- struct{}{}

	entity := event.Entity
	entity.LastSeen = event.Timestamp

	if err := monitorPtr.Store.UpdateEntity(entity); err != nil {
		logger.WithError(err).Error("error updating entity in store")
	}

	if err := monitorPtr.Store.UpdateKeepalive(event.Entity.Organization, event.Entity.ID, event.Timestamp+DefaultKeepaliveTimeout); err != nil {
		return err
	}

	return nil
}

// Stop the KeepaliveMonitor
func (monitorPtr *KeepaliveMonitor) Stop() {
	// atomically set stopped so that once Stop is called, all future
	// reads of stopped are true.
	if !atomic.CompareAndSwapInt32(&monitorPtr.stopped, 0, 1) {
		return
	}

	close(monitorPtr.reset)
}

// IsStopped returns true if the Monitor has been stopped.
func (monitorPtr *KeepaliveMonitor) IsStopped() bool {
	return atomic.LoadInt32(&monitorPtr.stopped) > 0
}
