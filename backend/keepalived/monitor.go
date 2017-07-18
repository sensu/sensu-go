package keepalived

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

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

	go func() {
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

				// test to see if the entity still exists (it may have been deleted)
				event, err := monitorPtr.Store.GetEventByEntityCheck(context.TODO(), monitorPtr.Entity.ID, "keepalive")
				if err != nil {
					// this should be a temporary error talking to the store. keep trying until
					// the store starts responding again.
					logger.WithError(err).Error("error getting keepalive event for client")
					break
				}

				// if the agent disconnected and reconnected elsewhere, stop the monitor
				// and return.
				if event != nil && event.Check.Status == 0 {
					monitorPtr.Stop()
					return
				}

				// if the entity is supposed to be deregistered, do so.
				if monitorPtr.Entity.Deregister {
					if err := monitorPtr.Deregisterer.Deregister(monitorPtr.Entity); err != nil {
						logger.WithError(err).Error("error deregistering entity")
					}
					monitorPtr.Stop()
					return
				}

				// this is a real keepalive event, emit it.
				if err := monitorPtr.EventCreator.Warn(monitorPtr.Entity); err != nil {
					logger.WithError(err).Error("error sending keepalive event")
				}
			}
			timer.Reset(timerDuration)
		}
	}()
}

// Update causes the KeepaliveMonitor to observe the event.
func (monitorPtr *KeepaliveMonitor) Update(event *types.Event) error {
	monitorPtr.reset <- struct{}{}

	entity := event.Entity
	entity.LastSeen = event.Timestamp
	ctx := context.WithValue(context.Background(), types.OrganizationKey, entity.Organization)

	if err := monitorPtr.Store.UpdateEntity(ctx, entity); err != nil {
		logger.WithError(err).Error("error updating entity in store")
	}

	prevEvent, err := monitorPtr.Store.GetEventByEntityCheck(ctx, entity.ID, "keepalive")
	if err != nil {
		logger.WithError(err).Error("error getting previous event from store")
	}

	if prevEvent != nil && prevEvent.Check.Status != 0 {
		monitorPtr.EventCreator.Resolve(entity)
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

// Reset the monitor's timer to emit an event at a given time.
func (monitorPtr *KeepaliveMonitor) Reset(t int64) {
	d := time.Duration(t - time.Now().Unix())
	if d < 0 {
		d = 0
	}
	monitorPtr.timer.Reset(d)
}
