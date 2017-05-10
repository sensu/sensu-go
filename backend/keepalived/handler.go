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

func (k *Keepalived) processKeepalives() {
	defer k.wg.Done()

	event := &types.Event{}

	for {
		select {
		case msg, ok := <-k.keepaliveChan:
			if ok {
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

				if err := k.Store.UpdateEntity(entity); err != nil {
					logger.WithError(err).Error("error updating entity in store")
					continue
				}

				if err := k.Store.UpdateKeepalive(event.Entity.ID, event.Timestamp+DefaultKeepaliveTimeout); err != nil {
					logger.WithError(err).Error("error updating keepalive in store")
					continue
				}
			}
		case <-k.stopping:
			return
		}
	}
}
