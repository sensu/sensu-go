package keepalived

import (
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// A Monitor observes events and takes actions based on them.
type Monitor interface {
	Update(event *types.Event) error
	Start()
	Stop()
}

// KeepaliveMonitor is a managed timer that is reset whenever the monitor
// observes a Keepalive event via the Update() function.
type KeepaliveMonitor struct {
	Entity       *types.Entity
	Deregisterer Deregisterer
	EventCreator EventCreator
	Store        store.Store

	reset    chan interface{}
	timer    *time.Timer
	stopping bool
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
				if monitorPtr.stopping {
					return
				}
				timer.Reset(timerDuration)

			case <-timer.C:
				// timed out keepalive
				if monitorPtr.Entity.Deregister {
					if err := monitorPtr.Deregisterer.Deregister(monitorPtr.Entity); err != nil {
						logger.WithError(err).Error("error deregistering entity")
					}
				} else {
					if err := monitorPtr.EventCreator.Warn(monitorPtr.Entity); err != nil {
						logger.WithError(err).Error("error sending keepalive event")
					}
				}

				timer.Reset(timerDuration)
			}
		}
	}()
	<-running
}

// Update causes the KeepaliveMonitor to observe the event.
func (monitorPtr *KeepaliveMonitor) Update(event *types.Event) error {
	if err := monitorPtr.Store.UpdateKeepalive(event.Entity.Organization, event.Entity.ID, event.Timestamp+DefaultKeepaliveTimeout); err != nil {
		return err
	}

	monitorPtr.reset <- struct{}{}
	return nil
}

// Stop the KeepaliveMonitor
func (monitorPtr *KeepaliveMonitor) Stop() {
	monitorPtr.stopping = true
	close(monitorPtr.reset)
}
