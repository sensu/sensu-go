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
		case msg := <-k.keepaliveChan:
			err := json.Unmarshal(msg, event)
			if err != nil {
				logger.WithError(err).Error("error unmarshaling keepliave event")
				continue
			}

			err = k.KeepaliveStore.UpdateKeepalive(event.Entity.ID, event.Timestamp, event.Timestamp+DefaultKeepaliveTimeout)
			if err != nil {
				logger.WithError(err).Error("error updating keepalive in store")
				continue
			}
		case <-k.stopping:
			return
		}
	}
}
