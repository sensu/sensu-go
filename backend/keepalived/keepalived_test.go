package keepalived

import (
	"context"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	stor "github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	storv2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockclientv3"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type mockDeregisterer struct {
	mock.Mock
}

type keepalivedTest struct {
	Keepalived    *Keepalived
	MessageBus    messaging.MessageBus
	StoreV2       *storetest.Store
	Store         *mockstore.MockStore
	Deregisterer  *mockDeregisterer
	receiver      chan interface{}
	SilencedCache cache.Cache
	Client        mockclientv3.MockClientV3
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

func (fakeLivenessInterface) BuryAndRevokeLease(context.Context, string) error {
	return nil
}

// type assertion
var _ liveness.Interface = fakeLivenessInterface{}

func fakeFactory(name string, dead, alive liveness.EventFunc, logger logrus.FieldLogger) liveness.Interface {
	return fakeLivenessInterface{}
}

func newKeepalivedTest(t *testing.T, client mockclientv3.MockClientV3) *keepalivedTest {
	store := &mockstore.MockStore{}
	storv2 := &storetest.Store{}
	deregisterer := &mockDeregisterer{}
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	k, err := New(Config{
		Client:          client,
		Store:           store,
		StoreV2:         storv2,
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
		StoreV2:      storv2,
		Deregisterer: deregisterer,
		Keepalived:   k,
		receiver:     make(chan interface{}),
		Client:       client,
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
			client := mockclientv3.MockClientV3{}
			getResp := &clientv3.GetResponse{}
			client.On("Get", mock.Anything, "/sensu.io/silenced/", mock.Anything).
				Return(getResp, nil)

			test := newKeepalivedTest(t, client)
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
	client := mockclientv3.MockClientV3{}
	getResp := &clientv3.GetResponse{}
	client.On("Get", mock.Anything, "/sensu.io/silenced/", mock.Anything).
		Return(getResp, nil)

	test := newKeepalivedTest(t, client)
	test.Store.On("GetFailingKeepalives", mock.Anything).Return([]*corev2.KeepaliveRecord{}, nil)
	require.NoError(t, test.Keepalived.Start())

	event := corev2.FixtureEvent("entity", "keepalive")
	event.Check.Status = 1

	test.StoreV2.On("CreateOrUpdate", mock.MatchedBy(func(req storv2.ResourceRequest) bool {
		return req.StoreName == new(corev3.EntityState).StoreName()
	}), mock.Anything).Return(nil)

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
	type assertionFunc func(*storetest.Store)

	newEntityWithClass := func(class string) *corev2.Entity {
		entity := corev2.FixtureEntity("agent1")
		entity.EntityClass = class
		return entity
	}

	newEntityConfigWithClass := func(class string) storv2.Wrapper {
		entity := corev3.FixtureEntityConfig("agent1")
		entity.EntityClass = class
		e, err := storv2.WrapResource(entity)
		require.NoError(t, err)
		return e
	}

	newAgentManagedEntity := func(class string) *corev2.Entity {
		entity := corev2.FixtureEntity("agent1")
		entity.EntityClass = corev2.EntityAgentClass
		entity.Labels = map[string]string{
			corev2.ManagedByLabel: "sensu-agent",
		}
		return entity
	}

	newAgentManagedEntityConfig := func(class string) storv2.Wrapper {
		entity := corev3.FixtureEntityConfig("agent1")
		entity.EntityClass = class
		entity.Metadata.Labels = map[string]string{
			corev2.ManagedByLabel: "sensu-agent",
		}
		e, err := storv2.WrapResource(entity)
		require.NoError(t, err)
		return e
	}

	firstSequenceEvent := &corev2.Event{Sequence: 1}

	tt := []struct {
		name                   string
		entity                 *corev2.Entity
		event                  *corev2.Event
		storeEntity            storv2.Wrapper
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
			storeCreateOrUpdateErr: &stor.ErrAlreadyExists{},
		},
		{
			name:             "Non-Registered Entity",
			entity:           newEntityWithClass("agent"),
			storeGetErr:      &stor.ErrNotFound{},
			event:            new(corev2.Event),
			expectedEventLen: 1,
		},
		{
			name:             "agent-managed entity is registered",
			entity:           newAgentManagedEntity("agent"),
			storeGetErr:      &stor.ErrNotFound{},
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
			assertionFunc: func(store *storetest.Store) {
				store.AssertCalled(t, "UpdateIfExists", mock.Anything, mock.Anything)
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			messageBus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
			require.NoError(t, err)
			require.NoError(t, messageBus.Start())

			storv2 := &storetest.Store{}

			tsubEvent := testSubscriber{
				ch: make(chan interface{}, 1),
			}
			subscriptionEvent, err := messageBus.Subscribe(messaging.TopicEvent, "testSubscriberEvent", tsubEvent)
			require.NoError(t, err)

			client := mockclientv3.MockClientV3{}
			getResp := &clientv3.GetResponse{}
			client.On("Get", mock.Anything, "/sensu.io/silenced/", mock.Anything).
				Return(getResp, nil)

			keepalived, err := New(Config{
				Client:          client,
				StoreV2:         storv2,
				Bus:             messageBus,
				LivenessFactory: fakeFactory,
				WorkerCount:     1,
				BufferSize:      1,
				StoreTimeout:    time.Minute,
			})
			require.NoError(t, err)

			storv2.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(tc.storeCreateOrUpdateErr)
			storv2.On("UpdateIfExists", mock.Anything, mock.Anything).Return(tc.storeCreateOrUpdateErr)
			storv2.On("Get", mock.Anything).Return(tc.storeEntity, tc.storeGetErr)
			err = keepalived.handleEntityRegistration(tc.entity, tc.event)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedEventLen, len(tsubEvent.ch))
			assert.NoError(t, subscriptionEvent.Cancel())

			if tc.assertionFunc != nil {
				tc.assertionFunc(storv2)
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
	store := &storetest.Store{}
	store.On("Get", mock.MatchedBy(func(req storv2.ResourceRequest) bool {
		return req.StoreName == new(corev3.EntityConfig).StoreName()
	})).Return((storv2.Wrapper)(nil), &stor.ErrNotFound{Key: "foo"})

	client := mockclientv3.MockClientV3{}
	getResp := &clientv3.GetResponse{}
	client.On("Get", mock.Anything, "/sensu.io/silenced/", mock.Anything).
		Return(getResp, nil)

	keepalived, err := New(Config{
		Client:          client,
		StoreV2:         store,
		Bus:             messageBus,
		LivenessFactory: fakeFactory,
		WorkerCount:     1,
		BufferSize:      1,
		StoreTimeout:    time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

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

	wrapper, err := storv2.WrapResource(
		corev3.FixtureEntityConfig("entity1"),
		[]wrap.Option{wrap.CompressNone, wrap.EncodeJSON}...)
	require.NoError(t, err)

	store := &storetest.Store{}
	store.On("Get", mock.MatchedBy(func(req storv2.ResourceRequest) bool {
		return req.StoreName == new(corev3.EntityConfig).StoreName()
	})).Return(wrapper, nil)
	store.On("Get", mock.MatchedBy(func(req storv2.ResourceRequest) bool {
		return req.StoreName == new(corev2.Event).StorePrefix()
	})).Return((*wrap.Wrapper)(nil), &stor.ErrNotFound{Key: "foo"})

	eventStore := &mockstore.MockStore{}
	eventStore.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), nil)

	client := mockclientv3.MockClientV3{}
	getResp := &clientv3.GetResponse{}
	client.On("Get", mock.Anything, "/sensu.io/silenced/", mock.Anything).
		Return(getResp, nil)

	keepalived, err := New(Config{
		Client:          client,
		StoreV2:         store,
		EventStore:      eventStore,
		Bus:             messageBus,
		LivenessFactory: fakeFactory,
		WorkerCount:     1,
		BufferSize:      1,
		StoreTimeout:    time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The switch should be buried since the event is nil
	if got, want := keepalived.dead("default/testSubscriber", liveness.Alive, true), true; got != want {
		t.Fatalf("got bury: %v, want bury: %v", got, want)
	}
}
