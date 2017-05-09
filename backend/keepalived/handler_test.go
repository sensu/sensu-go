package keepalived

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
)

func TestInvalidJSON(t *testing.T) {
	store := &mockstore.MockStore{}

	event1 := types.FixtureEvent("entity1", "check")
	store.On("UpdateKeepalive", event1.Entity.ID, event1.Timestamp+DefaultKeepaliveTimeout).Return(nil)

	event2 := types.FixtureEvent("entity2", "check")
	store.On("UpdateKeepalive", event2.Entity.ID, event2.Timestamp+DefaultKeepaliveTimeout).Return(errors.New(""))

	k := &Keepalived{}

	messageBus := &messaging.WizardBus{}
	messageBus.Start()
	defer messageBus.Stop()
	k.MessageBus = messageBus
	k.KeepaliveStore = store
	k.Start()
	k.keepaliveChan <- []byte(".")

	event1Bytes, _ := json.Marshal(event1)
	k.keepaliveChan <- event1Bytes

	event2Bytes, _ := json.Marshal(event2)
	k.keepaliveChan <- event2Bytes

	k.Stop()
}
