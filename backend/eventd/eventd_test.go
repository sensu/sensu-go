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

func TestEventFlappingHandling(t *testing.T) {
	// Mock eventd
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

	// Mock calls to the store
	var nilEvent *types.Event
	mockStore.On(
		"GetEventByEntityCheck",
		mock.Anything,
		"foo",
		"check_foo",
	).Return(nilEvent, nil)

	mockStore.On("UpdateEvent", mock.AnythingOfType("*types.Event")).Return(nil)

	// Mock an event message
	event := types.FixtureEvent("foo", "check_foo")
	event.Check.Config.HighFlapThreshold = 30
	event.Check.Config.LowFlapThreshold = 10
	event.Check.History = fictionalHistory()

	// Mock the message handling
	err = e.handleMessage(event)
	assert.NoError(t, err)

	// Make sure the event has been marked as flapping
	assert.Equal(t, types.EventFlappingAction, event.Check.Action)
}

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
		mock.Anything,
		"entity",
		"check",
	).Return(nilEvent, nil)
	mockStore.On("UpdateEvent", mock.AnythingOfType("*types.Event")).Return(nil)

	bus.Publish(messaging.TopicEventRaw, event)

	err = e.Stop()
	assert.NoError(t, err)

	mockStore.AssertCalled(t, "UpdateEvent", mock.AnythingOfType("*types.Event"))

	// Make sure the event has been marked as created
	assert.Equal(t, types.EventCreateAction, event.Check.Action)
}
