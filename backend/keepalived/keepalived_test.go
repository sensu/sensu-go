package keepalived

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/monitor"
	"github.com/sensu/sensu-go/testing/mockring"
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
	receiver     chan interface{}
}

func (k *keepalivedTest) Receiver() chan<- interface{} {
	return k.receiver
}

type fakeMonitorSupervisor struct {
}

func (f fakeMonitorSupervisor) Monitor(context.Context, string, *types.Event, int64) error {
	return nil
}

func fakeFactory(monitor.Handler) monitor.Supervisor {
	return fakeMonitorSupervisor{}
}

// type assertion
var _ monitor.Supervisor = fakeMonitorSupervisor{}

func newKeepalivedTest(t *testing.T) *keepalivedTest {
	store := &mockstore.MockStore{}
	deregisterer := &mockDeregisterer{}
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
		RingGetter: &mockring.Getter{},
	})
	require.NoError(t, err)
	k, err := New(Config{Store: store, Bus: bus, MonitorFactory: fakeFactory})
	require.NoError(t, err)
	test := &keepalivedTest{
		MessageBus:   bus,
		Store:        store,
		Deregisterer: deregisterer,
		Keepalived:   k,
		receiver:     make(chan interface{}),
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

			k := test.Keepalived

			test.Store.On("GetFailingKeepalives", mock.Anything).Return(tc.records, nil)
			for _, event := range tc.events {
				test.Store.On("GetEventByEntityCheck", mock.Anything, event.Entity.ID, "keepalive").Return(event, nil)
				if event.Check.Status != 0 {
					test.Store.On("UpdateFailingKeepalive", mock.Anything, event.Entity, mock.AnythingOfType("int64")).Return(nil)
				}
			}

			require.NoError(t, k.Start())

			var err error
			select {
			case err = <-k.Err():
			default:
			}
			assert.NoError(t, err)
			assert.NoError(t, k.Stop())
		})
	}
}

func TestEventProcessing(t *testing.T) {
	test := newKeepalivedTest(t)
	test.Store.On("GetFailingKeepalives", mock.Anything).Return([]*types.KeepaliveRecord{}, nil)
	require.NoError(t, test.Keepalived.Start())
	event := types.FixtureEvent("entity", "keepalive")
	event.Check.Status = 1

	test.Store.On("UpdateEntity", mock.Anything, event.Entity).Return(nil)
	test.Store.On("DeleteFailingKeepalive", mock.Anything, event.Entity).Return(nil)

	test.Keepalived.keepaliveChan <- event
	assert.NoError(t, test.Keepalived.Stop())
}

type testSubscriber struct {
	ch chan interface{}
}

func (t testSubscriber) Receiver() chan<- interface{} {
	return t.ch
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
			messageBus, err := messaging.NewWizardBus(messaging.WizardBusConfig{
				RingGetter: &mockring.Getter{},
			})
			require.NoError(t, err)
			require.NoError(t, messageBus.Start())

			store := &mockstore.MockStore{}

			tsub := testSubscriber{
				ch: make(chan interface{}, 1),
			}
			subscription, err := messageBus.Subscribe(messaging.TopicEvent, "testSubscriber", tsub)
			require.NoError(t, err)

			keepalived, err := New(Config{Store: store, Bus: messageBus, MonitorFactory: fakeFactory})
			require.NoError(t, err)

			store.On("GetEntityByID", mock.Anything, "agent1").Return(tc.storeEntity, nil)
			err = keepalived.handleEntityRegistration(tc.entity)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedLen, len(tsub.ch))
			assert.NoError(t, subscription.Cancel())
		})
	}
}

func TestCreateKeepaliveEvent(t *testing.T) {
	event := types.FixtureEvent("entity1", "keepalive")
	keepaliveEvent := createKeepaliveEvent(event)
	assert.Equal(t, "keepalive", keepaliveEvent.Check.Name)
	assert.Equal(t, uint32(60), keepaliveEvent.Check.Interval)
	assert.Equal(t, []string{"keepalive"}, keepaliveEvent.Check.Handlers)
	assert.Equal(t, uint32(0), keepaliveEvent.Check.Status)
	assert.NotEqual(t, int64(0), keepaliveEvent.Check.Issued)

	event.Check = nil
	keepaliveEvent = createKeepaliveEvent(event)
	assert.Equal(t, "keepalive", keepaliveEvent.Check.Name)
	assert.Equal(t, uint32(20), keepaliveEvent.Check.Interval)
	assert.Equal(t, uint32(120), keepaliveEvent.Check.Timeout)
}
