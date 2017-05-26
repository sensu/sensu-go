package keepalived

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type mockMessageBus struct {
	mock.Mock
}

func (m *mockMessageBus) Subscribe(topic, consumer string, channel chan<- interface{}) error {
	args := m.Called(topic, consumer, channel)
	return args.Error(0)
}
func (m *mockMessageBus) Unsubscribe(topic, consumer string) error {
	args := m.Called(topic, consumer)
	return args.Error(0)
}
func (m *mockMessageBus) Publish(topic string, message interface{}) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

func (m *mockMessageBus) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockMessageBus) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockMessageBus) Status() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockMessageBus) Err() <-chan error {
	errChan := make(chan error, 1)
	return errChan
}

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type HandlerTestSuite struct {
	suite.Suite
	store            *mockstore.MockStore
	bus              *mockMessageBus
	keepalived       *Keepalived
	stoppingMonitors chan struct{}
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (s *HandlerTestSuite) SetupTest() {
	s.store = &mockstore.MockStore{}
	s.bus = &mockMessageBus{}
	s.stoppingMonitors = make(chan struct{})

	s.keepalived = &Keepalived{
		MessageBus: s.bus,
		Store:      s.store,
		stopping:   make(chan struct{}, 1),
	}
}

func (s *HandlerTestSuite) TestKeepaliveUpdates() {
	event := types.FixtureEvent("entity1", "check1")

	keepaliveUpdated := make(chan struct{})

	s.store.On("UpdateKeepalive", event.Entity.ID, mock.AnythingOfType("int64")).Return(nil).Run(func(args mock.Arguments) { close(keepaliveUpdated) })

	ch := make(chan *types.Event)

	go s.keepalived.monitorEntity(ch, event.Entity, s.stoppingMonitors)

	ch <- event

	s.NotNil(<-keepaliveUpdated)

	close(s.keepalived.stopping)
}

func (s *HandlerTestSuite) TestKeepaliveTimeout() {
	keepaliveTimeout = 1
	keepaliveTimedout := make(chan struct{})

	s.bus.On("Publish", messaging.TopicEvent, mock.AnythingOfType("[]uint8")).Return(nil).Run(func(args mock.Arguments) { close(keepaliveTimedout) })

	ch := make(chan *types.Event)

	entity := types.FixtureEntity("entity1")

	go s.keepalived.monitorEntity(ch, entity, s.stoppingMonitors)

	s.NotNil(<-keepaliveTimedout)

	close(s.keepalived.stopping)
}

func (s *HandlerTestSuite) TestKeepaliveTimeoutDeregister() {
	keepaliveTimeout = 1
	entityDeleted := make(chan struct{})
	eventPublished := make(chan struct{})
	eventDeleted := make(chan struct{})

	event := types.FixtureEvent("entity1", "check1")
	event.Entity.Deregister = true

	mockEvents := []*types.Event{
		event,
	}

	s.store.On("DeleteEntity", event.Entity).Return(nil).Run(func(args mock.Arguments) { close(entityDeleted) })
	s.store.On("GetEventsByEntity", event.Entity.ID).Return(mockEvents, nil)
	s.store.On("DeleteEventByEntityCheck", event.Entity.ID, event.Check.Name).Return(nil).Run(func(args mock.Arguments) { close(eventDeleted) })
	s.bus.On("Publish", messaging.TopicEvent, mock.AnythingOfType("[]uint8")).Return(nil).Run(func(args mock.Arguments) { close(eventPublished) })

	ch := make(chan *types.Event)

	go s.keepalived.monitorEntity(ch, event.Entity, s.stoppingMonitors)

	s.NotNil(<-entityDeleted)
	s.NotNil(<-eventDeleted)
	s.NotNil(<-eventPublished)

	close(s.keepalived.stopping)
}

func (s *HandlerTestSuite) TestKeepaliveTimeoutDeregistrationHandler() {
	keepaliveTimeout = 1
	eventPublished := make(chan struct{})

	event := types.FixtureEvent("entity1", "check1")
	event.Entity.Deregister = true
	event.Entity.Deregistration = types.Deregistration{
		Handler: "deregistration",
	}

	mockEvents := []*types.Event{
		event,
	}

	s.store.On("DeleteEntity", event.Entity).Return(nil)
	s.store.On("GetEventsByEntity", event.Entity.ID).Return(mockEvents, nil)
	s.store.On("DeleteEventByEntityCheck", event.Entity.ID, event.Check.Name).Return(nil)
	s.bus.On("Publish", messaging.TopicEvent, mock.AnythingOfType("[]uint8")).Return(nil).Run(func(args mock.Arguments) {
		publishedEvent := types.Event{}
		_ = json.Unmarshal(args[1].([]byte), &publishedEvent)

		if publishedEvent.Check.Name == "deregistration" {
			s.Equal([]string{"deregistration"}, publishedEvent.Check.Handlers)
			close(eventPublished)
		}
	})

	ch := make(chan *types.Event)

	go s.keepalived.monitorEntity(ch, event.Entity, s.stoppingMonitors)

	s.NotNil(<-eventPublished)

	close(s.keepalived.stopping)
}

// Run the HandlerTestSuite
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
