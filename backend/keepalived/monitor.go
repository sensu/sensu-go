package keepalived

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// KeepaliveMonitor is a managed timer that is reset whenever the monitor
// observes a Keepalive event via the Update() function. Once the timer has
// been stopped, it cannot be started or used again.
type KeepaliveMonitor struct {
	Entity       *types.Entity
	Deregisterer Deregisterer
	EventCreator EventCreator
	MessageBus   messaging.MessageBus
	Store        store.Store

	reset   chan interface{}
	timer   *time.Timer
	stopped int32
	failing int32
}

// Start initializes the monitor and starts its monitoring goroutine.
// If the monitor has been previously stopped, this method has no
// effect.
func (monitorPtr *KeepaliveMonitor) Start() {
	if monitorPtr.IsStopped() {
		return
	}

	timerDuration := time.Duration(monitorPtr.Entity.KeepaliveTimeout) * time.Second
	monitorPtr.timer = time.NewTimer(timerDuration)
	monitorPtr.reset = make(chan interface{})
	go func() {
		timer := monitorPtr.timer
		ctx := types.SetContextFromResource(context.Background(), monitorPtr.Entity)

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
				// timed out keepalive

				// if the entity is supposed to be deregistered, do so.
				if monitorPtr.Entity.Deregister {
					if err = monitorPtr.Deregisterer.Deregister(monitorPtr.Entity); err != nil {
						logger.WithError(err).Error("error deregistering entity")
					}
					monitorPtr.Stop()
					return
				}

				// this is a real keepalive event, emit it.
				if err = monitorPtr.EventCreator.Warn(monitorPtr.Entity); err != nil {
					logger.WithError(err).Error("error sending keepalive event")
				}

				timeout = time.Now().Unix() + int64(monitorPtr.Entity.KeepaliveTimeout)
				if err = monitorPtr.Store.UpdateFailingKeepalive(ctx, monitorPtr.Entity, timeout); err != nil {
					logger.WithError(err).Error("error updating failing keepalive in store")
				}

				atomic.CompareAndSwapInt32(&monitorPtr.failing, 0, 1)
			}

			if monitorPtr.IsStopped() {
				return
			}

			timer.Reset(timerDuration)
		}
	}()
}

// Update causes the KeepaliveMonitor to observe the event. If the monitor has
// been stopped, this method has no effect.
func (monitorPtr *KeepaliveMonitor) Update(event *types.Event) error {
	// once the monitor is stopped, we can't continue, because the
	// reset channel will be closed.
	if monitorPtr.IsStopped() {
		return nil
	}

	entity := event.Entity

	if atomic.CompareAndSwapInt32(&monitorPtr.failing, 1, 0) {
		if err := monitorPtr.Store.DeleteFailingKeepalive(context.Background(), entity); err != nil {
			logger.Debug(err)
		}
	}

	monitorPtr.reset <- struct{}{}

	entity.LastSeen = event.Timestamp
	ctx := types.SetContextFromResource(context.Background(), entity)

	if err := monitorPtr.Store.UpdateEntity(ctx, entity); err != nil {
		logger.WithError(err).Error("error updating entity in store")
	}

	return monitorPtr.EventCreator.Pass(entity)
}

// Stop the KeepaliveMonitor. Once the monitor has been stopped it
// cannot be used any longer.
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
// Once the monitor has been stopped, this has no effect.
func (monitorPtr *KeepaliveMonitor) Reset(t int64) {
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
