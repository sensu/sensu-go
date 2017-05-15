package keepalived

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultKeepaliveTimeout is the amount of time we consider a Keepalive
	// valid for.
	DefaultKeepaliveTimeout = 120 // seconds
)

var keepaliveTimeout = 0

// if this function returns, it is because keepalived is shutting down
func (k *Keepalived) startWorkers() {
	entityChannels := map[string](chan *types.Event){}

	// concurrent access to entityChannels map is problematic
	// because of race conditions :(
	k.HandlerCount = 1

	k.wg = &sync.WaitGroup{}
	k.wg.Add(k.HandlerCount)

	for i := 0; i < k.HandlerCount; i++ {
		go k.processKeepalives(entityChannels)
	}
}

func (k *Keepalived) processKeepalives(ec map[string](chan *types.Event)) {
	defer k.wg.Done()

	var (
		channel chan *types.Event
		ok      bool
	)

	stoppingMonitors := make(chan struct{})

	for {
		select {
		case msg, open := <-k.keepaliveChan:
			if open {
				event := &types.Event{}
				if err := json.Unmarshal(msg, event); err != nil {
					logger.WithError(err).Error("error unmarshaling keepalive event")
					continue
				}

				entity := event.Entity
				if err := entity.Validate(); err != nil {
					logger.WithError(err).Error("invalid keepalive event")
					continue
				}
				entity.LastSeen = event.Timestamp

				if err := k.Store.UpdateEntity(entity); err != nil {
					logger.WithError(err).Error("error updating entity in store")
					continue
				}

				channel, ok = ec[entity.ID]
				if !ok {
					channel = make(chan *types.Event)
					ec[entity.ID] = channel
					go k.monitorEntity(channel, entity, stoppingMonitors)
				}

				channel <- event
			}
		case <-k.stopping:
			close(stoppingMonitors)
			return
		}
	}
}

func (k *Keepalived) deregisterEntity(entity *types.Entity) {
	if err := k.Store.DeleteEntity(entity); err != nil {
		logger.WithError(err).Error("error deleting entity in store")
	}
	events, err := k.Store.GetEventsByEntity(entity.ID)
	if err != nil {
		logger.WithError(err).Error("error fetching events for entity")
	}
	for _, event := range events {
		// Delete event from Store
		if err := k.Store.DeleteEventByEntityCheck(entity.ID, event.Check.Name); err != nil {
			logger.WithError(err).Error("error deleting event for entity")
		}
		event.Check.Output = "Resolving due to entity deregistering"
		event.Check.Status = 0
		event.Check.History = []types.CheckHistory{}

		eventBytes, err := json.Marshal(event)
		if err != nil {
			logger.Errorf("error serializing deregistration event: %s", err.Error())
		}
		k.MessageBus.Publish(messaging.TopicEvent, eventBytes)
	}
}

func (k *Keepalived) createKeepaliveEvent(entity *types.Entity) {
	keepaliveCheck := &types.Check{
		Name:          "keepalive",
		Interval:      DefaultKeepaliveTimeout,
		Subscriptions: []string{""},
		Command:       "",
		Handlers:      []string{"keepalive"},
	}
	keepaliveEvent := &types.Event{
		Entity: entity,
		Check:  keepaliveCheck,
	}
	eventBytes, err := json.Marshal(keepaliveEvent)
	if err != nil {
		logger.Errorf("error serializing keepalive event: %s", err.Error())
	}
	k.MessageBus.Publish(messaging.TopicEvent, eventBytes)
}

func (k *Keepalived) monitorEntity(ch chan *types.Event, entity *types.Entity, stoppingMonitors chan struct{}) {
	timeout := DefaultKeepaliveTimeout
	if keepaliveTimeout != 0 {
		timeout = keepaliveTimeout
	}
	timerDuration := time.Duration(timeout) * time.Second
	timer := time.NewTimer(timerDuration)

	for {
		select {
		case event := <-ch:
			if err := k.Store.UpdateKeepalive(event.Entity.ID, event.Timestamp+DefaultKeepaliveTimeout); err != nil {
				logger.WithError(err).Error("error updating keepalive in store")
				continue
			}

			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(timerDuration)
		case <-timer.C:
			// timed out keepalive
			if entity.Deregister {
				k.deregisterEntity(entity)
			} else {
				k.createKeepaliveEvent(entity)
			}
		case <-stoppingMonitors:
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}
}
