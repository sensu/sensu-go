package eventd

import (
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

	bus.Publish(messaging.TopicEventRaw, nil)

	badEvent := &types.Event{}
	badEvent.Check = &types.Check{}
	badEvent.Entity = &types.Entity{}
	badEvent.Timestamp = time.Now().Unix()

	bus.Publish(messaging.TopicEventRaw, badEvent)

	event := types.FixtureEvent("entity", "check")

	var nilEvent *types.Event
	// no previous event.
	mockStore.On(
		"GetEventByEntityCheck",
		mock.AnythingOfType("*context.valueCtx"),
		"entity",
		"check",
	).Return(nilEvent, nil)
	mockStore.On("UpdateEvent", mock.AnythingOfType("*types.Event")).Return(nil)

	bus.Publish(messaging.TopicEventRaw, event)

	err = e.Stop()
	assert.NoError(t, err)

	mockStore.AssertCalled(t, "UpdateEvent", mock.AnythingOfType("*types.Event"))
}
