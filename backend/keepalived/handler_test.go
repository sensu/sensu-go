package keepalived

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

func TestInvalidJSON(t *testing.T) {
	store := &mockstore.MockStore{}

	event1 := types.FixtureEvent("entity1", "check")
	store.On("UpdateEntity", mock.AnythingOfType("*types.Entity")).Return(nil)
	store.On("UpdateKeepalive", event1.Entity.Organization, event1.Entity.ID, event1.Timestamp+DefaultKeepaliveTimeout).Return(nil)

	event2 := types.FixtureEvent("entity2", "check")
	store.On("UpdateKeepalive", event2.Entity.Organization, event2.Entity.ID, event2.Timestamp+DefaultKeepaliveTimeout).Return(errors.New(""))

	k := &Keepalived{}

	messageBus := &messaging.WizardBus{}
	messageBus.Start()
	defer messageBus.Stop()
	k.MessageBus = messageBus
	k.Store = store
	k.Start()
	k.keepaliveChan <- []byte(".")
	k.keepaliveChan <- event1
	k.keepaliveChan <- event2

	k.Stop()
}
