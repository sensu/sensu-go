package keepalived

import (
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including assertion methods.
type HandlerTestSuite struct {
	suite.Suite
	store            *mockstore.MockStore
	bus              *mockbus.MockBus
	keepalived       *Keepalived
	stoppingMonitors chan struct{}
}

// Make sure that VariableThatShouldStartAtFive is set to five
// before each test
func (s *HandlerTestSuite) SetupTest() {
	s.store = &mockstore.MockStore{}
	s.bus = &mockbus.MockBus{}
	s.bus.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.stoppingMonitors = make(chan struct{})

	s.keepalived = &Keepalived{
		MessageBus: s.bus,
		Store:      s.store,
		stopping:   make(chan struct{}, 1),
	}
	s.keepalived.Start()
}

func (s *HandlerTestSuite) TestKeepaliveUpdates() {
	event := types.FixtureEvent("entity1", "check1")

	keepaliveUpdated := make(chan struct{})

	s.store.On("UpdateKeepalive", event.Entity.Organization, event.Entity.ID, mock.AnythingOfType("int64")).Return(nil).Run(func(args mock.Arguments) { close(keepaliveUpdated) })

	ch := make(chan *types.Event)

	go s.keepalived.monitorEntity(ch, event.Entity, s.stoppingMonitors)

	ch <- event

	s.NotNil(<-keepaliveUpdated)

	close(s.keepalived.stopping)
}

func (s *HandlerTestSuite) TestKeepaliveTimeout() {
	keepaliveTimeout = 1
	keepaliveTimedout := make(chan struct{})

	s.bus.On("Publish", messaging.TopicEvent, mock.Anything).Return(nil).Run(func(args mock.Arguments) { close(keepaliveTimedout) })

	ch := make(chan *types.Event)

	entity := types.FixtureEntity("entity1")

	go s.keepalived.monitorEntity(ch, entity, s.stoppingMonitors)

	s.NotNil(<-keepaliveTimedout)

	close(s.keepalived.stopping)
}

// Run the HandlerTestSuite
func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
