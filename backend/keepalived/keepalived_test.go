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
)

type mockDeregisterer struct {
	mock.Mock
}

type keepalivedTest struct {
	Keepalived   *Keepalived
	MessageBus   messaging.MessageBus
	Store        *mockstore.MockStore
	Deregisterer *mockDeregisterer
}

func newKeepalivedTest(t *testing.T) *keepalivedTest {
	store := &mockstore.MockStore{}
	deregisterer := &mockDeregisterer{}
	bus := &messaging.WizardBus{}
	test := &keepalivedTest{
		MessageBus:   bus,
		Store:        store,
		Deregisterer: deregisterer,
		Keepalived: &Keepalived{
			Store:      store,
			MessageBus: bus,
			MonitorFactory: func(*types.Entity, *types.Event, time.Duration, monitor.UpdateHandler, monitor.FailureHandler) monitor.Interface {
				mon := &mockmonitor.MockMonitor{}
				mon.On("HandleUpdate", mock.Anything).Return(nil)
				return mon
			},
		},
	}
	require.NoError(t, test.MessageBus.Start())
	return test
}

func (k *keepalivedTest) Dispose(t *testing.T) {
	assert.NoError(t, k.MessageBus.Stop())
}

func TestStartStop(t *testing.T) {
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

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			test := newKeepalivedTest(t)
			defer test.Dispose(t)

			k := &Keepalived{}
			assert.Error(t, k.Start())

			k.MessageBus = test.MessageBus
			assert.Error(t, k.Start())

			store := &mockstore.MockStore{}
			store.On("GetFailingKeepalives", mock.Anything).Return(tc.records, nil)
			for _, event := range tc.events {
				store.On("GetEventByEntityCheck", mock.Anything, event.Entity.ID, "keepalive").Return(event, nil)
				if event.Check.Status != 0 {
					store.On("UpdateFailingKeepalive", mock.Anything, event.Entity, mock.AnythingOfType("int64")).Return(nil)
				}
			}

			k.Store = store
			assert.NoError(t, k.Start())
			assert.NotNil(t, k.MonitorFactory, "*Keepalived.Start() ensures there is a MonitorFactory")

			assert.NoError(t, k.Status())

			var err error
			select {
			case err = <-k.Err():
			default:
			}
			assert.NoError(t, err)
			assert.NoError(t, k.Stop())
			assert.Equal(t, tc.monitors, len(k.monitors))
		})
	}
}

func TestEventProcessing(t *testing.T) {
	test := newKeepalivedTest(t)
	test.Store.On("GetFailingKeepalives", mock.Anything).Return([]*types.KeepaliveRecord{}, nil)
	mon := &mockmonitor.MockMonitor{}
	mon.On("HandleUpdate", mock.Anything).Return(nil)
	test.Keepalived.MonitorFactory = func(entity *types.Entity, event *types.Event, t time.Duration, updateHandler monitor.UpdateHandler, failureHandler monitor.FailureHandler) monitor.Interface {
		return mon
	}
	require.NoError(t, test.Keepalived.Start())
	event := types.FixtureEvent("entity", "keepalive")
	event.Check.Status = 1

	test.Store.On("UpdateEntity", mock.Anything, event.Entity).Return(nil)
	test.Store.On("DeleteFailingKeepalive", mock.Anything, event.Entity).Return(nil)

	test.Keepalived.keepaliveChan <- event
	assert.NoError(t, test.Keepalived.Stop())
	mon.AssertCalled(t, "HandleUpdate", event)
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
