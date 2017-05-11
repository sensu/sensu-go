package keepalived

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

const (
	// DefaultKeepaliveTimeout is the amount of time we consider a Keepalive
	// valid for.
	DefaultKeepaliveTimeout = 120 // seconds
)

// if this function returns, it is because keepalived is shutting down
func (k *Keepalived) processKeepalives() {
	defer k.wg.Done()

	event := &types.Event{}
	var (
		channel chan *types.Event
		ok      bool
	)
	entityChannels := map[string](chan *types.Event){}

	for {
		select {
		case msg, open := <-k.keepaliveChan:
			if open {
				if err := json.Unmarshal(msg, event); err != nil {
					logger.WithError(err).Error("error unmarshaling keepliave event")
					continue
				}

				entity := event.Entity
				if err := entity.Validate(); err != nil {
					logger.WithError(err).Error("invalid keepalive event")
					continue
				}
				entity.LastSeen = event.Timestamp

				channel, ok = entityChannels[entity.ID]
				if !ok {
					channel = make(chan *types.Event)
					entityChannels[entity.ID] = channel
					go k.monitorEntity(channel)
				}

				channel <- event

				if err := k.Store.UpdateEntity(entity); err != nil {
					logger.WithError(err).Error("error updating entity in store")
					continue
				}
			}
		case <-k.stopping:
			return
		}
	}
}

func (k *Keepalived) monitorEntity(ch chan *types.Event) {
	for {
		select {
		case event := <-ch:
			if err := k.Store.UpdateKeepalive(event.Entity.ID, event.Timestamp+DefaultKeepaliveTimeout); err != nil {
				logger.WithError(err).Error("error updating keepalive in store")
				continue
			}
		}
	}
}
