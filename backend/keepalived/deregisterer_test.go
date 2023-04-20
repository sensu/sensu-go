package keepalived

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockcache"
	"github.com/sensu/sensu-go/testing/mockstore"
)

func TestDeregister(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.V2MockStore{}
	mockEventStore := &mockstore.MockStore{}
	mockBus := &mockbus.MockBus{}

	adapter := &Deregistration{
		Store:		mockStore,
		MessageBus:	mockBus,
	}

	entity := corev2.FixtureEntity("entity")
	entity.Deregister = true
	check := corev2.FixtureCheck("check")
	event := corev2.FixtureEvent(entity.Name, check.Name)

	ecstore := new(mockstore.EntityConfigStore)
	mockStore.On("GetEntityConfigStore").Return(ecstore)
	mockStore.On("GetEventStore").Return(mockEventStore)

	ecstore.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockEventStore.On("GetEventsByEntity", mock.Anything, entity.Name, &store.SelectionPredicate{}).Return([]*corev2.Event{event}, nil)
	mockEventStore.On("DeleteEventByEntityCheck", mock.Anything, entity.Name, check.Name).Return(nil)

	mockBus.On("Publish", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	assert.NoError(adapter.Deregister(entity))
}

func TestDeregistrationHandler(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.V2MockStore{}
	mockEventStore := &mockstore.MockStore{}
	ecstore := new(mockstore.EntityConfigStore)
	mockStore.On("GetEventStore").Return(mockEventStore)
	mockStore.On("GetEntityConfigStore").Return(ecstore)

	mockBus := &mockbus.MockBus{}
	mockCache := &mockcache.MockCache{}

	adapter := &Deregistration{
		Store:		mockStore,
		MessageBus:	mockBus,
		SilencedCache:	mockCache,
	}

	entity := corev2.FixtureEntity("entity")
	entity.Deregister = true
	entity.Deregistration = corev2.Deregistration{
		Handler: "deregistration",
	}
	check := corev2.FixtureCheck("check")

	mockCache.On("Get", "default").Once().Return(
		[]cachev2.Value[*corev2.Silenced, corev2.Silenced]{
			{Resource: corev2.FixtureSilenced("*:deregistration")},
		},
	)

	ecstore.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockEventStore.On("GetEventsByEntity", mock.Anything, entity.Name, &store.SelectionPredicate{}).Return([]*corev2.Event{}, nil)
	mockEventStore.On("DeleteEventByEntityCheck", mock.Anything, entity.Name, check.Name).Return(nil)

	mockBus.On("Publish", messaging.TopicEvent, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		event := args[1].(*corev2.Event)
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
