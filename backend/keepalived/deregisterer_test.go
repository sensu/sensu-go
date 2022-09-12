package keepalived

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockcache"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
)

func TestDeregister(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.V2MockStore{}
	mockEventStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}

	adapter := &Deregistration{
		EntityStore: mockStore,
		EventStore:  mockEventStore,
		MessageBus:  mockBus,
	}

	entity := types.FixtureEntity("entity")
	entity.Deregister = true
	check := types.FixtureCheck("check")
	event := types.FixtureEvent(entity.Name, check.Name)

	mockStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	mockEventStore.On("GetEventsByEntity", mock.Anything, entity.Name, &store.SelectionPredicate{}).Return([]*types.Event{event}, nil)
	mockEventStore.On("DeleteEventByEntityCheck", mock.Anything, entity.Name, check.Name).Return(nil)

	mockBus.On("Publish", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	assert.NoError(adapter.Deregister(entity))
}

func TestDeregistrationHandler(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.V2MockStore{}
	mockEventStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}
	mockCache := &mockcache.MockCache{}

	adapter := &Deregistration{
		EntityStore:   mockStore,
		EventStore:    mockEventStore,
		MessageBus:    mockBus,
		SilencedCache: mockCache,
	}

	entity := types.FixtureEntity("entity")
	entity.Deregister = true
	entity.Deregistration = types.Deregistration{
		Handler: "deregistration",
	}
	check := types.FixtureCheck("check")

	mockCache.On("Get", "default").Once().Return(
		[]cache.Value{
			{Resource: corev2.FixtureSilenced("*:deregistration")},
		},
	)

	mockStore.On("Delete", mock.Anything, mock.Anything).Return(nil)

	mockEventStore.On("GetEventsByEntity", mock.Anything, entity.Name, &store.SelectionPredicate{}).Return([]*types.Event{}, nil)
	mockEventStore.On("DeleteEventByEntityCheck", mock.Anything, entity.Name, check.Name).Return(nil)

	mockBus.On("Publish", messaging.TopicEvent, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		event := args[1].(*types.Event)
		assert.Equal("deregistration", event.Entity.Deregistration.Handler)
		assert.Equal("deregistration", event.Check.Name)
		assert.Equal(0, len(event.Check.Subscriptions))
		assert.Equal(true, event.IsSilenced(), "event is not silenced")
		if event.Timestamp == 0 {
			t.Fatal("event timestamp is nil, expected a timestamp in the deregistration event")
		}
		if event.GetUUID() == uuid.Nil {
			t.Fatal("event UUID is nil")
		}
	})

	assert.NoError(adapter.Deregister(entity))
}
