package adapters

import (
	"testing"

	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoDeregister(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}

	adapter := &Deregistration{
		Store:      mockStore,
		MessageBus: mockBus,
	}

	entity := types.FixtureEntity("entity1")

	assert.NoError(adapter.Deregister(entity))
}

func TestDeregister(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}

	adapter := &Deregistration{
		Store:      mockStore,
		MessageBus: mockBus,
	}

	entity := types.FixtureEntity("entity")
	check := types.FixtureCheck("check")
	event := types.FixtureEvent(entity.ID, check.Config.Name)

	mockStore.On("GetEventsByEntity", entity.Organization, entity.ID).Return([]*types.Event{event}, nil)
	mockStore.On("DeleteEvent", event).Return(nil)
	mockStore.On("DeleteEntity", entity).Return(nil)

	mockBus.On("Publish", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	assert.NoError(adapter.Deregister(entity))
}
