package keepalived

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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

type fakeLivenessInterface struct {
}

func (fakeLivenessInterface) Alive(context.Context, string, int64) error {
	return nil
}

func (fakeLivenessInterface) Dead(context.Context, string, int64) error {
	return nil
}

func (fakeLivenessInterface) Bury(context.Context, string) error {
	return nil
}

// type assertion
var _ liveness.Interface = fakeLivenessInterface{}

func fakeFactory(name string, dead, alive liveness.EventFunc, logger logrus.FieldLogger) liveness.Interface {
	return fakeLivenessInterface{}
}

func newKeepalivedTest(t *testing.T) *keepalivedTest {
	store := &mockstore.MockStore{}
	deregisterer := &mockDeregisterer{}
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	k, err := New(Config{
		Store:           store,
		EventStore:      store,
		Bus:             bus,
		LivenessFactory: fakeFactory,
		BufferSize:      1,
		WorkerCount:     1,
		StoreTimeout:    time.Second,
	})
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
	failingEvent := func(e *corev2.Event) *corev2.Event {
		e.Check.Status = 1
		return e
	}

	tt := []struct {
		name     string
		records  []*corev2.KeepaliveRecord
		events   []*corev2.Event
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
			records: []*corev2.KeepaliveRecord{
				{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "entity1",
						Namespace: "org",
					},
					Time: 0,
				},
				{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "entity2",
						Namespace: "org",
					},
					Time: 0,
				},
			},
			events: []*corev2.Event{
				corev2.FixtureEvent("entity1", "keepalive"),
				corev2.FixtureEvent("entity2", "keepalive"),
			},
			monitors: 0,
		},
		{
			name: "Failing Keepalives",
			records: []*corev2.KeepaliveRecord{
				{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "entity1",
						Namespace: "org",
					},
					Time: 0,
				},
				{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "entity2",
						Namespace: "org",
					},
					Time: 0,
				},
			},
			events: []*corev2.Event{
				failingEvent(corev2.FixtureEvent("entity1", "keepalive")),
				failingEvent(corev2.FixtureEvent("entity2", "keepalive")),
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
				test.Store.On("GetEventByEntityCheck", mock.Anything, event.Entity.Name, "keepalive").Return(event, nil)
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
	test.Store.On("GetFailingKeepalives", mock.Anything).Return([]*corev2.KeepaliveRecord{}, nil)
	require.NoError(t, test.Keepalived.Start())
	event := corev2.FixtureEvent("entity", "keepalive")
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
	newEntityWithClass := func(class string) *corev2.Entity {
		entity := corev2.FixtureEntity("agent1")
		entity.EntityClass = class
		return entity
	}

	tt := []struct {
		name        string
		entity      *corev2.Entity
		storeEntity *corev2.Entity
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
			messageBus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
			require.NoError(t, err)
			require.NoError(t, messageBus.Start())

			store := &mockstore.MockStore{}

			tsub := testSubscriber{
				ch: make(chan interface{}, 1),
			}
			subscription, err := messageBus.Subscribe(messaging.TopicEvent, "testSubscriber", tsub)
			require.NoError(t, err)

			keepalived, err := New(Config{
				Store:           store,
				EventStore:      store,
				Bus:             messageBus,
				LivenessFactory: fakeFactory,
				WorkerCount:     1,
				BufferSize:      1,
				StoreTimeout:    time.Minute,
			})
			require.NoError(t, err)

			store.On("GetEntityByName", mock.Anything, "agent1").Return(tc.storeEntity, nil)
			err = keepalived.handleEntityRegistration(tc.entity)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedLen, len(tsub.ch))
			assert.NoError(t, subscription.Cancel())
		})
	}
}

func TestCreateKeepaliveEvent(t *testing.T) {
	event := corev2.FixtureEvent("entity1", "keepalive")
	keepaliveEvent := createKeepaliveEvent(event)
	assert.Equal(t, "keepalive", keepaliveEvent.Check.Name)
	assert.Equal(t, uint32(60), keepaliveEvent.Check.Interval)
	assert.Equal(t, []string{"keepalive"}, keepaliveEvent.Check.Handlers)
	assert.Equal(t, uint32(0), keepaliveEvent.Check.Status)
	assert.Equal(t, "default", keepaliveEvent.Check.Namespace)
	assert.Equal(t, "default", keepaliveEvent.ObjectMeta.Namespace)
	assert.NotEqual(t, int64(0), keepaliveEvent.Check.Issued)

	event.Check = nil
	keepaliveEvent = createKeepaliveEvent(event)
	assert.Equal(t, "keepalive", keepaliveEvent.Check.Name)
	assert.Equal(t, uint32(20), keepaliveEvent.Check.Interval)
	assert.Equal(t, uint32(120), keepaliveEvent.Check.Timeout)
}

func TestCreateRegistrationEvent(t *testing.T) {
	event := corev2.FixtureEntity("entity1")
	keepaliveEvent := createRegistrationEvent(event)
	assert.Equal(t, RegistrationCheckName, keepaliveEvent.Check.Name)
	assert.Equal(t, uint32(1), keepaliveEvent.Check.Interval)
	assert.Equal(t, []string{RegistrationHandlerName}, keepaliveEvent.Check.Handlers)
	assert.Equal(t, uint32(1), keepaliveEvent.Check.Status)
	assert.Equal(t, "default", keepaliveEvent.Check.Namespace)
	assert.Equal(t, "default", keepaliveEvent.ObjectMeta.Namespace)
}

func TestDeadCallbackNoEntity(t *testing.T) {
	messageBus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := messageBus.Start(); err != nil {
		t.Fatal(err)
	}
	tsub := testSubscriber{
		ch: make(chan interface{}, 1),
	}
	if _, err := messageBus.Subscribe(messaging.TopicEvent, "testSubscriber", tsub); err != nil {
		t.Fatal(err)
	}
	store := &mockstore.MockStore{}
	keepalived, err := New(Config{
		Store:           store,
		EventStore:      store,
		Bus:             messageBus,
		LivenessFactory: fakeFactory,
		WorkerCount:     1,
		BufferSize:      1,
		StoreTimeout:    time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	store.On("GetEntityByName", mock.Anything, mock.Anything).Return((*corev2.Entity)(nil), nil)

	if got, want := keepalived.dead("default/testSubscriber", liveness.Alive, true), true; got != want {
		t.Fatalf("got bury: %v, want bury: %v", got, want)
	}
}

func TestDeadCallbackNoEvent(t *testing.T) {
	messageBus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := messageBus.Start(); err != nil {
		t.Fatal(err)
	}
	tsub := testSubscriber{
		ch: make(chan interface{}, 1),
	}
	if _, err := messageBus.Subscribe(messaging.TopicEvent, "testSubscriber", tsub); err != nil {
		t.Fatal(err)
	}
	store := &mockstore.MockStore{}
	keepalived, err := New(Config{
		Store:           store,
		EventStore:      store,
		Bus:             messageBus,
		LivenessFactory: fakeFactory,
		WorkerCount:     1,
		BufferSize:      1,
		StoreTimeout:    time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	store.On("GetEntityByName", mock.Anything, mock.Anything).Return(corev2.FixtureEntity("foo"), nil)
	store.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), nil)

	// The switch should be buried since the event is nil
	if got, want := keepalived.dead("default/testSubscriber", liveness.Alive, true), true; got != want {
		t.Fatalf("got bury: %v, want bury: %v", got, want)
	}
}
