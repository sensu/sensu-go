package keepalived

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type KeepalivedTestSuite struct {
	suite.Suite
	Keepalived   *Keepalived
	MessageBus   messaging.MessageBus
	Store        *mockstore.MockStore
	Deregisterer *mockDeregisterer
	EventCreator *mockCreator
}

func (suite *KeepalivedTestSuite) SetupTest() {
	suite.MessageBus = &messaging.WizardBus{}
	suite.MessageBus.Start()

	mockStore := &mockstore.MockStore{}
	dereg := &mockDeregisterer{}
	creator := &mockCreator{}

	suite.Deregisterer = dereg
	suite.EventCreator = creator
	suite.Store = mockStore

	keepalived := &Keepalived{
		Store:      suite.Store,
		MessageBus: suite.MessageBus,
		MonitorFactory: func(e *types.Entity) *KeepaliveMonitor {
			return &KeepaliveMonitor{
				Entity:       e,
				Deregisterer: dereg,
				EventCreator: creator,
				Store:        mockStore,
			}
		},
	}

	suite.Keepalived = keepalived
}

func (suite *KeepalivedTestSuite) AfterTest() {
	suite.MessageBus.Stop()
	suite.Keepalived.Stop()
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

			k.MonitorFactory = nil

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
	suite.Keepalived.MonitorFactory = nil
	suite.NoError(suite.Keepalived.Start())
	event := types.FixtureEvent("entity", "keepalive")
	event.Check.Status = 1

	suite.Store.On("UpdateEntity", mock.Anything, event.Entity).Return(nil)

	suite.Store.On("GetEventByEntityCheck", mock.Anything, event.Entity.ID, "keepalive").Return(event, nil)
	suite.Keepalived.keepaliveChan <- event
	time.Sleep(100 * time.Millisecond)
	suite.Store.AssertCalled(suite.T(), "UpdateEntity", mock.Anything, event.Entity)
}

func TestKeepalivedSuite(t *testing.T) {
	suite.Run(t, new(KeepalivedTestSuite))
}
