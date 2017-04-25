package eventd

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEventHandling(t *testing.T) {
	bus := &messaging.WizardBus{}
	bus.Start()

	mockStore := &mockstore.MockStore{}

	e := &Eventd{
		Store:        mockStore,
		MessageBus:   bus,
		HandlerCount: 5,
	}

	err := e.Start()
	assert.NoError(t, err)

	bus.Publish(messaging.TopicEventRaw, []byte("{}"))

	badEvent := &types.Event{}
	badEvent.Check = &types.Check{}
	badEvent.Entity = &types.Entity{}
	badEvent.Timestamp = time.Now().Unix()

	eventBytes, _ := json.Marshal(badEvent)
	bus.Publish(messaging.TopicEventRaw, eventBytes)

	event := types.FixtureEvent("entity", "check")

	eventBytes, _ = json.Marshal(event)

	var nilEvent *types.Event
	// no previous event.
	mockStore.On("GetEventByEntityCheck", "entity", "check").Return(nilEvent, nil)
	// We can't directly mock this, because we don't have access to the event that
	// will be deserialized, so we mock it with a function pointer instead.
	mockStore.On("UpdateEvent", mock.AnythingOfType("*types.Event")).Return(nil)

	bus.Publish(messaging.TopicEventRaw, eventBytes)

	err = e.Stop()
	assert.NoError(t, err)
}
