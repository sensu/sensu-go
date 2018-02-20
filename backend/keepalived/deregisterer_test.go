package keepalived

import (
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeregister(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}

	adapter := &Deregistration{
		Store:      mockStore,
		MessageBus: mockBus,
	}

	entity := types.FixtureEntity("entity")
	entity.Deregister = true
	check := types.FixtureCheck("check")
	event := types.FixtureEvent(entity.ID, check.Name)

	mockStore.On("GetEventsByEntity", mock.Anything, entity.ID).Return([]*types.Event{event}, nil)
	mockStore.On("DeleteEventByEntityCheck", mock.Anything, entity.ID, check.Name).Return(nil)
	mockStore.On("DeleteEntity", mock.Anything, entity).Return(nil)

	mockBus.On("Publish", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	assert.NoError(adapter.Deregister(entity))
}

func TestDeregistrationHandler(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}

	adapter := &Deregistration{
		Store:      mockStore,
		MessageBus: mockBus,
	}

	entity := types.FixtureEntity("entity")
	entity.Deregister = true
	entity.Deregistration = types.Deregistration{
		Handler: "deregistration",
	}
	check := types.FixtureCheck("check")

	mockStore.On("GetEventsByEntity", mock.Anything, entity.ID).Return([]*types.Event{}, nil)
	mockStore.On("DeleteEventByEntityCheck", mock.Anything, entity.ID, check.Name).Return(nil)
	mockStore.On("DeleteEntity", mock.Anything, entity).Return(nil)

	mockBus.On("Publish", messaging.TopicEvent, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		event := args[1].(*types.Event)
		assert.Equal("deregistration", event.Entity.Deregistration.Handler)
	})

	assert.NoError(adapter.Deregister(entity))
}
