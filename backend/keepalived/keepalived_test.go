package keepalived

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/testing/mockmonitor"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type mockDeregisterer struct {
	mock.Mock
}

type KeepalivedTestSuite struct {
	suite.Suite
	Keepalived   *Keepalived
	MessageBus   messaging.MessageBus
	Store        *mockstore.MockStore
	Deregisterer *mockDeregisterer
}

func (suite *KeepalivedTestSuite) SetupTest() {
	suite.MessageBus = &messaging.WizardBus{}
	suite.NoError(suite.MessageBus.Start())

	mockStore := &mockstore.MockStore{}
	dereg := &mockDeregisterer{}

	suite.Deregisterer = dereg
	suite.Store = mockStore

	keepalived := &Keepalived{
		Store:      suite.Store,
		MessageBus: suite.MessageBus,
	}

	keepalived.MonitorFactory = func(*types.Entity, time.Duration, monitor.UpdateHandler, monitor.FailureHandler) monitor.Interface {
		mon := &mockmonitor.MockMonitor{}
		mon.On("HandleUpdate", mock.Anything).Return(nil)
		return mon
	}

	suite.Keepalived = keepalived
}

func (suite *KeepalivedTestSuite) AfterTest() {
	suite.NoError(suite.MessageBus.Stop())
	suite.NoError(suite.Keepalived.Stop())
}

func (suite *KeepalivedTestSuite) TestStartStop() {
	failingEvent := func(e *types.Event) *types.Event {
		e.Check.Status = 1
		return e
	}

	tt := []struct {
		name     string
		records  []*types.KeepaliveRecord
		events   []*types.Event
		monitors int
	}{
		{
			name:     "No Keepalives",
			records:  nil,
			events:   nil,
			monitors: 0,
		},
		{
			name: "Passing Keepalives",
			records: []*types.KeepaliveRecord{
				{
					EntityID:     "entity1",
					Organization: "org",
					Time:         0,
				},
				{
					EntityID:     "entity2",
					Organization: "org",
					Time:         0,
				},
			},
			events: []*types.Event{
				types.FixtureEvent("entity1", "keepalive"),
				types.FixtureEvent("entity2", "keepalive"),
			},
			monitors: 0,
		},
		{
			name: "Failing Keepalives",
			records: []*types.KeepaliveRecord{
				{
					EntityID:     "entity1",
					Organization: "org",
					Time:         0,
				},
				{
					EntityID:     "entity2",
					Organization: "org",
					Time:         0,
				},
			},
			events: []*types.Event{
				failingEvent(types.FixtureEvent("entity1", "keepalive")),
				failingEvent(types.FixtureEvent("entity2", "keepalive")),
			},
			monitors: 2,
		},
	}

	t := suite.T()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			k := &Keepalived{}
			suite.Error(k.Start())

			k.MessageBus = suite.MessageBus
			suite.Error(k.Start())

			store := &mockstore.MockStore{}
			store.On("GetFailingKeepalives", mock.Anything).Return(tc.records, nil)
			for _, event := range tc.events {
				store.On("GetEventByEntityCheck", mock.Anything, event.Entity.ID, "keepalive").Return(event, nil)
				if event.Check.Status != 0 {
					store.On("UpdateFailingKeepalive", mock.Anything, event.Entity, mock.AnythingOfType("int64")).Return(nil)
				}
			}

			k.Store = store
			suite.NoError(k.Start())
			suite.NotNil(k.MonitorFactory, "*Keepalived.Start() ensures there is a MonitorFactory")

			suite.NoError(k.Status())

			var err error
			select {
			case err = <-k.Err():
			default:
			}
			suite.NoError(err)

			suite.NoError(k.Stop())

			suite.Equal(tc.monitors, len(k.monitors))
		})
	}

}

func (suite *KeepalivedTestSuite) TestEventProcessing() {
	suite.Store.On("GetFailingKeepalives", mock.Anything).Return([]*types.KeepaliveRecord{}, nil)
	mon := &mockmonitor.MockMonitor{}
	mon.On("HandleUpdate", mock.Anything).Return(nil)
	suite.Keepalived.MonitorFactory = func(e *types.Entity, t time.Duration, updateHandler monitor.UpdateHandler, failureHandler monitor.FailureHandler) monitor.Interface {
		return mon
	}
	suite.NoError(suite.Keepalived.Start())
	event := types.FixtureEvent("entity", "keepalive")
	event.Check.Status = 1

	suite.Store.On("UpdateEntity", mock.Anything, event.Entity).Return(nil)
	suite.Store.On("DeleteFailingKeepalive", mock.Anything, event.Entity).Return(nil)

	suite.Keepalived.keepaliveChan <- event
	suite.NoError(suite.Keepalived.Stop())
	mon.AssertCalled(suite.T(), "HandleUpdate", event)
}

func TestKeepalivedSuite(t *testing.T) {
	suite.Run(t, new(KeepalivedTestSuite))
}

func TestProcessRegistration(t *testing.T) {
	newEntityWithClass := func(class string) *types.Entity {
		entity := types.FixtureEntity("agent1")
		entity.Class = class
		return entity
	}

	tt := []struct {
		name        string
		entity      *types.Entity
		storeEntity *types.Entity
		expectedLen int
	}{
		{
			name:        "Registered Entity Without Agent Class",
			entity:      newEntityWithClass("router"),
			storeEntity: newEntityWithClass("router"),
			expectedLen: 0,
		},
		{
			name:        "Registered Entity With Agent Class",
			entity:      newEntityWithClass("agent"),
			storeEntity: newEntityWithClass("agent"),
			expectedLen: 0,
		},
		{
			name:        "Non-Registered Entity With Agent Class",
			entity:      newEntityWithClass("agent"),
			storeEntity: nil,
			expectedLen: 1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			messageBus := &messaging.WizardBus{}
			require.NoError(t, messageBus.Start())

			store := &mockstore.MockStore{}

			testChan := make(chan interface{}, 1)
			err := messageBus.Subscribe(messaging.TopicEvent, "test-subscriber", testChan)
			require.NoError(t, err)

			keepalived := &Keepalived{
				Store:      store,
				MessageBus: messageBus,
			}

			store.On("GetEntityByID", mock.Anything, "agent1").Return(tc.storeEntity, nil)
			err = keepalived.handleEntityRegistration(tc.entity)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedLen, len(testChan))
		})
	}
}
