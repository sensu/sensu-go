package keepalived

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
)

type KeepaliveStore struct {
	mock.Mock
}

func (s *KeepaliveStore) DeleteFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig) error {
	args := s.Called(ctx, entityConfig)
	return args.Error(0)
}

func (s *KeepaliveStore) GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*corev2.KeepaliveRecord), args.Error(1)
}

func (s *KeepaliveStore) UpdateFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig, expiration int64) error {
	args := s.Called(ctx, entityConfig, expiration)
	return args.Error(0)
}

type mockDeregisterer struct {
	mock.Mock
}

type mockOperatorMonitor struct {
	mock.Mock
}

func (m *mockOperatorMonitor) MonitorOperators(ctx context.Context, req store.MonitorOperatorsRequest) <-chan []store.OperatorState {
	return m.Called(ctx, req).Get(0).(<-chan []store.OperatorState)
}

type mockOperatorConcierge struct {
	mock.Mock
}

func (m *mockOperatorConcierge) CheckIn(ctx context.Context, state store.OperatorState) error {
	return m.Called(ctx, state).Error(0)
}

func (m *mockOperatorConcierge) CheckOut(ctx context.Context, key store.OperatorKey) error {
	return m.Called(ctx, key).Error(0)
}

type keepalivedTest struct {
	Keepalived        *Keepalived
	MessageBus        messaging.MessageBus
	Store             *mockstore.V2MockStore
	EventStore        *mockstore.MockStore
	KeepaliveStore    *KeepaliveStore
	Deregisterer      *mockDeregisterer
	receiver          chan interface{}
	SilencedCache     SilencesCache
	OperatorMonitor   *mockOperatorMonitor
	OperatorConcierge *mockOperatorConcierge
}

func (k *keepalivedTest) Receiver() chan<- interface{} {
	return k.receiver
}

func newKeepalivedTest(t *testing.T) *keepalivedTest {
	stor := new(mockstore.V2MockStore)
	eventStore := &mockstore.MockStore{}
	ec := new(mockstore.EntityConfigStore)
	es := new(mockstore.EntityStateStore)
	stor.On("GetEventStore").Return(eventStore)
	stor.On("GetEntityStore").Return(eventStore)
	stor.On("GetEntityConfigStore").Return(ec)
	stor.On("GetEntityStateStore").Return(es)
	ec.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	es.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	keepaliveStore := &KeepaliveStore{}
	deregisterer := &mockDeregisterer{}
	bus, _ := messaging.NewWizardBus(messaging.WizardBusConfig{})
	operatorMonitor := new(mockOperatorMonitor)
	operatorMonitor.On("MonitorOperators", mock.Anything, mock.Anything).Return(make(<-chan []store.OperatorState))
	operatorConcierge := new(mockOperatorConcierge)
	operatorConcierge.On("CheckIn", mock.Anything, mock.Anything).Return(nil)
	k, err := New(Config{
		Store:             stor,
		EventStore:        eventStore,
		Bus:               bus,
		BufferSize:        1,
		WorkerCount:       1,
		StoreTimeout:      time.Second,
		OperatorMonitor:   operatorMonitor,
		OperatorConcierge: operatorConcierge,
	})
	if err != nil {
		panic(err)
	}
	k.reconstructionPeriod = 0
	test := &keepalivedTest{
		MessageBus:        bus,
		Store:             stor,
		EventStore:        eventStore,
		KeepaliveStore:    keepaliveStore,
		Deregisterer:      deregisterer,
		Keepalived:        k,
		receiver:          make(chan interface{}),
		OperatorMonitor:   operatorMonitor,
		OperatorConcierge: operatorConcierge,
	}
	if err := test.MessageBus.Start(); err != nil {
		panic(err)
	}
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

			test.KeepaliveStore.On("GetFailingKeepalives", mock.Anything).Return(tc.records, nil)
			for _, event := range tc.events {
				test.EventStore.On("GetEventByEntityCheck", mock.Anything, event.Entity.Name, "keepalive").Return(event, nil)
				if event.Check.Status != 0 {
					test.KeepaliveStore.On("UpdateFailingKeepalive", mock.Anything, event.Entity, mock.AnythingOfType("int64")).Return(nil)
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
	test.KeepaliveStore.On("GetFailingKeepalives", mock.Anything).Return([]*corev2.KeepaliveRecord{}, nil)
	require.NoError(t, test.Keepalived.Start())

	event := corev2.FixtureEvent("entity", "keepalive")
	event.Check.Status = 1

	test.EventStore.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	test.KeepaliveStore.On("DeleteFailingKeepalive", mock.Anything, mock.Anything).Return(nil)

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
	type assertionFunc func(*mockstore.V2MockStore)

	newEntityWithClass := func(class string) *corev2.Entity {
		entity := corev2.FixtureEntity("agent1")
		entity.EntityClass = class
		return entity
	}

	newEntityConfigWithClass := func(class string) *corev3.EntityConfig {
		entity := corev3.FixtureEntityConfig("agent1")
		entity.EntityClass = class
		return entity
	}

	newAgentManagedEntity := func(class string) *corev2.Entity {
		entity := corev2.FixtureEntity("agent1")
		entity.EntityClass = corev2.EntityAgentClass
		entity.Labels = map[string]string{
			corev2.ManagedByLabel: "sensu-agent",
		}
		return entity
	}

	newAgentManagedEntityConfig := func(class string) *corev3.EntityConfig {
		entity := corev3.FixtureEntityConfig("agent1")
		entity.EntityClass = class
		entity.Metadata.Labels = map[string]string{
			corev2.ManagedByLabel: "sensu-agent",
		}
		return entity
	}

	firstSequenceEvent := &corev2.Event{Sequence: 1}

	tt := []struct {
		name                   string
		entity                 *corev2.Entity
		event                  *corev2.Event
		storeEntity            *corev3.EntityConfig
		expectedEventLen       int
		storeGetErr            error
		storeCreateOrUpdateErr error
		expectedEntityLen      int
		assertionFunc          assertionFunc
	}{
		{
			name:                   "Registered Entity",
			entity:                 newEntityWithClass("agent"),
			storeEntity:            newEntityConfigWithClass("agent"),
			event:                  new(corev2.Event),
			expectedEventLen:       0,
			storeCreateOrUpdateErr: &store.ErrAlreadyExists{},
		},
		{
			name:             "Non-Registered Entity",
			entity:           newEntityWithClass("agent"),
			storeGetErr:      &store.ErrNotFound{},
			event:            new(corev2.Event),
			expectedEventLen: 1,
		},
		{
			name:             "agent-managed entity is registered",
			entity:           newAgentManagedEntity("agent"),
			storeGetErr:      &store.ErrNotFound{},
			event:            firstSequenceEvent,
			expectedEventLen: 1,
		},
		{
			name:             "agent-managed entity config is updated on reconnect",
			entity:           newAgentManagedEntity("agent"),
			storeEntity:      newEntityConfigWithClass("agent"),
			event:            firstSequenceEvent,
			expectedEventLen: 0,
		},
		{
			name:             "backend-managed entity config no longer has the managed_by label with sensu-agent",
			entity:           newEntityWithClass("agent"),
			storeEntity:      newAgentManagedEntityConfig("agent"),
			event:            new(corev2.Event),
			expectedEventLen: 0,
			assertionFunc: func(store *mockstore.V2MockStore) {
				store.GetEntityConfigStore().(*mockstore.EntityConfigStore).AssertCalled(t, "UpdateIfExists", mock.Anything, mock.Anything)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			messageBus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
			require.NoError(t, err)
			require.NoError(t, messageBus.Start())

			store := &mockstore.V2MockStore{}

			tsubEvent := testSubscriber{
				ch: make(chan interface{}, 1),
			}
			subscriptionEvent, err := messageBus.Subscribe(messaging.TopicEvent, "testSubscriberEvent", tsubEvent)
			require.NoError(t, err)

			keepalived, err := New(Config{
				Store:        store,
				Bus:          messageBus,
				WorkerCount:  1,
				BufferSize:   1,
				StoreTimeout: time.Minute,
			})
			require.NoError(t, err)

			ec := new(mockstore.EntityConfigStore)
			es := new(mockstore.EntityStateStore)
			store.On("GetEntityConfigStore").Return(ec)
			store.On("GetEntityStateStore").Return(es)

			ec.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(tc.storeCreateOrUpdateErr)
			es.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(tc.storeCreateOrUpdateErr)
			ec.On("UpdateIfExists", mock.Anything, mock.Anything, mock.Anything).Return(tc.storeCreateOrUpdateErr)
			es.On("UpdateIfExists", mock.Anything, mock.Anything, mock.Anything).Return(tc.storeCreateOrUpdateErr)
			ec.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(tc.storeEntity, tc.storeGetErr)
			err = keepalived.handleEntityRegistration(tc.entity, tc.event)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedEventLen, len(tsubEvent.ch))
			assert.NoError(t, subscriptionEvent.Cancel())

			if tc.assertionFunc != nil {
				tc.assertionFunc(store)
			}
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
	assert.Equal(t, corev2.RegistrationCheckName, keepaliveEvent.Check.Name)
	assert.Equal(t, uint32(1), keepaliveEvent.Check.Interval)
	assert.Equal(t, []string{corev2.RegistrationHandlerName}, keepaliveEvent.Check.Handlers)
	assert.Equal(t, uint32(1), keepaliveEvent.Check.Status)
	assert.Equal(t, "default", keepaliveEvent.Check.Namespace)
	assert.Equal(t, "default", keepaliveEvent.ObjectMeta.Namespace)
}
