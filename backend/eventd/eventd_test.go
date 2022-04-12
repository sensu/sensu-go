package eventd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockcache"
	"github.com/sensu/sensu-go/testing/mockstore"
)

type fakeSwitchSet struct{}

func (f *fakeSwitchSet) Alive(context.Context, string, int64) error {
	return nil
}

func (*fakeSwitchSet) Dead(context.Context, string, int64) error {
	return nil
}

func (*fakeSwitchSet) Bury(context.Context, string) error {
	return nil
}

func (*fakeSwitchSet) BuryAndRevokeLease(context.Context, string) error {
	return nil
}

func newFakeFactory(f liveness.Interface) liveness.Factory {
	return func(name string, dead, alive liveness.EventFunc, logger logrus.FieldLogger) liveness.Interface {
		return f
	}
}

func newEventd(store storev2.Interface, eventStore store.Store, bus messaging.MessageBus, livenessFactory liveness.Factory) *Eventd {
	ctx, cancel := context.WithCancel(context.Background())
	return &Eventd{
		ctx:             ctx,
		cancel:          cancel,
		store:           store,
		eventStore:      eventStore,
		bus:             bus,
		livenessFactory: livenessFactory,
		errChan:         make(chan error, 1),
		shutdownChan:    make(chan struct{}, 1),
		eventChan:       make(chan interface{}, 100),
		wg:              &sync.WaitGroup{},
		mu:              &sync.Mutex{},
		Logger:          NoopLogger{},
		workerCount:     5,
		storeTimeout:    time.Minute,
		silencedCache:   &cache.Resource{},
	}
}

func TestEventHandling(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockEntityStore := &storetest.Store{}
	mockStore := &mockstore.MockStore{}
	e := newEventd(mockEntityStore, mockStore, bus, newFakeFactory(&fakeSwitchSet{}))

	require.NoError(t, e.Start())
	require.NoError(t, bus.Publish(messaging.TopicEventRaw, nil))

	badEvent := &corev2.Event{}
	badEvent.Check = &corev2.Check{}
	badEvent.Entity = &corev2.Entity{}
	badEvent.Timestamp = time.Now().Unix()

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, badEvent))

	event := corev2.FixtureEvent("entity", "check")
	addMockEntityV2(t, mockEntityStore, event.Entity)

	var nilEvent *corev2.Event
	// no previous event.
	mockStore.On(
		"GetEventByEntityCheck",
		mock.Anything,
		"entity",
		"check",
	).Return(nilEvent, nil)
	event.Check.Occurrences = 1
	event.Check.State = corev2.EventPassingState
	event.Check.LastOK = event.Timestamp
	mockStore.On("UpdateEvent", mock.Anything).Return(event, nilEvent, nil)

	// No silenced entries
	mockStore.On(
		"GetSilencedEntriesBySubscription",
		mock.Anything,
	).Return([]*corev2.Silenced{}, nil)
	mockStore.On(
		"GetSilencedEntriesByCheckName",
		mock.Anything,
	).Return([]*corev2.Silenced{}, nil)

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, event))

	err = e.Stop()
	assert.NoError(t, err)

	mockStore.AssertCalled(t, "UpdateEvent", mock.Anything)

	assert.Equal(t, int64(1), event.Check.Occurrences)

	// Make sure the event has been marked with the proper state
	assert.Equal(t, corev2.EventPassingState, event.Check.State)
	assert.Equal(t, event.Timestamp, event.Check.LastOK)

}

func TestEventMonitor(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockEntityStore := &storetest.Store{}
	mockStore := &mockstore.MockStore{}
	e := newEventd(mockEntityStore, mockStore, bus, newFakeFactory(&fakeSwitchSet{}))

	require.NoError(t, e.Start())
	require.NoError(t, bus.Publish(messaging.TopicEventRaw, nil))

	event := corev2.FixtureEvent("entity", "check")
	event.Check.Ttl = 90

	addMockEntityV2(t, mockEntityStore, event.Entity)

	var nilEvent *corev2.Event
	// no previous event.
	mockStore.On(
		"GetEventByEntityCheck",
		mock.Anything,
		"entity",
		"check",
	).Return(nilEvent, nil)
	event.Check.State = corev2.EventPassingState
	mockStore.On("UpdateEvent", mock.Anything).Return(event, nilEvent, nil)

	// No silenced entries
	mockStore.On(
		"GetSilencedEntriesBySubscription",
		mock.Anything,
	).Return([]*corev2.Silenced{}, nil)
	mockStore.On(
		"GetSilencedEntriesByCheckName",
		mock.Anything,
	).Return([]*corev2.Silenced{}, nil)

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, event))

	err = e.Stop()
	assert.NoError(t, err)

	// Make sure the event has been marked with the proper state
	assert.Equal(t, corev2.EventPassingState, event.Check.State)
}

type mockSwitchSet struct {
	mock.Mock
}

func (m *mockSwitchSet) Alive(ctx context.Context, id string, ttl int64) error {
	args := m.Called(ctx, id, ttl)
	return args.Error(0)
}

func (m *mockSwitchSet) Dead(ctx context.Context, id string, ttl int64) error {
	args := m.Called(ctx, id, ttl)
	return args.Error(0)
}

func (m *mockSwitchSet) Bury(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSwitchSet) BuryAndRevokeLease(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCheckTTL(t *testing.T) {
	ttlCheck := corev2.FixtureEvent("entity", "check")
	ttlCheck.Check.Ttl = 120
	nonTTLCheck := corev2.FixtureEvent("entity", "check")

	type switchesFunc func(*mockSwitchSet)

	tests := []struct {
		name             string
		msg              interface{}
		previousEvent    *corev2.Event
		previousEventErr error
		switchesFunc     switchesFunc
		wantErr          bool
	}{
		{
			name:    "a check without TTL shouldn't bury its switchset",
			msg:     nonTTLCheck,
			wantErr: false,
		},
		{
			name: "a check with TTL should update its switchset with alive",
			msg:  ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Alive", mock.Anything, "default/check/entity", int64(120)).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "switchset alive err should be returned",
			msg:  ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Alive", mock.Anything, "default/check/entity", int64(120)).
					Return(errors.New("err"))
			},
			wantErr: true,
		},
		{
			name:          "a check that just got its TTL disabled should bury its switchet",
			msg:           nonTTLCheck,
			previousEvent: ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Bury", mock.Anything, "default/check/entity").Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "an error while burying the switchset should not be returned",
			msg:           nonTTLCheck,
			previousEvent: ttlCheck,
			switchesFunc: func(s *mockSwitchSet) {
				s.On("Bury", mock.Anything, "default/check/entity").Return(errors.New("err"))
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &storetest.Store{}
			eventStore := &mockstore.MockStore{}
			switches := &mockSwitchSet{}
			if tt.switchesFunc != nil {
				tt.switchesFunc(switches)
			}

			mockEvent := corev2.FixtureEvent("entity", "mock")
			addMockEntityV2(t, store, mockEvent.Entity)

			e := &Eventd{
				store:           store,
				eventStore:      eventStore,
				livenessFactory: newFakeFactory(switches),
				workerCount:     1,
				wg:              &sync.WaitGroup{},
				Logger:          NoopLogger{},
				silencedCache:   &cache.Resource{},
			}

			var err error
			e.bus, err = messaging.NewWizardBus(messaging.WizardBusConfig{})
			require.NoError(t, err)
			require.NoError(t, e.bus.Start())

			eventStore.On("GetEventByEntityCheck", mock.Anything, "entity", "check").
				Return(tt.previousEvent, tt.previousEventErr)
			eventStore.On("GetSilencedEntriesBySubscription", mock.Anything, mock.Anything).
				Return([]*corev2.Silenced{}, nil)
			eventStore.On("GetSilencedEntriesByCheckName", mock.Anything, mock.Anything).
				Return([]*corev2.Silenced{}, nil)
			eventStore.On("UpdateEvent", mock.Anything, mock.Anything).Return(tt.msg, mockEvent, nil)

			if _, err := e.handleMessage(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("Eventd.handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuryConditions(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		storeFunc      func(*storetest.Store)
		eventStoreFunc func(*mockstore.MockStore)
		bury           bool
	}{
		{
			// TODO: this test case should have bury: false when round robin check
			// TTL is supported
			name: "bury switch on round robin",
			key:  "default/foo",
			bury: true,
		},
		{
			name: "bury switch on nil entity",
			key:  "default/foo/bar",
			storeFunc: func(store *storetest.Store) {
				entityNotFound(store, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: true,
		},
		{
			name: "do not bury switch on entity lookup error",
			key:  "default/foo/bar",
			storeFunc: func(s *storetest.Store) {
				entityLookupError(s, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: false,
		},
		{
			name: "bury switch on nil event",
			key:  "default/foo/bar",
			storeFunc: func(s *storetest.Store) {
				addFixtureEntity(s, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return((*corev2.Event)(nil), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: true,
		},
		{
			name: "do not bury switch on event lookup error",
			key:  "default/foo/bar",
			storeFunc: func(s *storetest.Store) {
				addFixtureEntity(s, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return((*corev2.Event)(nil), errors.New("!"))
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: false,
		},
		{
			name: "do not bury switch otherwise",
			key:  "default/foo/bar",
			storeFunc: func(s *storetest.Store) {
				addFixtureEntity(s, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return(corev2.FixtureEvent("bar", "foo"), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: false,
		},
		{
			name: "bury when failing keepalive exists",
			key:  "default/foo/bar",
			storeFunc: func(s *storetest.Store) {
				addFixtureEntity(s, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(&corev2.Event{
					Check: &corev2.Check{
						Status: 1,
					},
				}, nil)
			},
			bury: true,
		},
		{
			name: "do not bury when passing keepalive exists",
			key:  "default/foo/bar",
			storeFunc: func(s *storetest.Store) {
				addFixtureEntity(s, "default", "bar")
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(&corev2.Event{
					Check: &corev2.Check{
						Status: 0,
					},
				}, nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return(corev2.FixtureEvent("bar", "foo"), nil)
			},
			bury: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := new(storetest.Store)
			if test.storeFunc != nil {
				test.storeFunc(store)
			}

			eventStore := new(mockstore.MockStore)
			if test.eventStoreFunc != nil {
				test.eventStoreFunc(eventStore)
			}

			eventd := &Eventd{store: store, eventStore: eventStore, ctx: context.Background()}
			if got, want := eventd.dead(test.key, liveness.Alive, false), test.bury; got != want {
				t.Fatalf("bad bury result: got %v, want %v", got, want)
			}
		})
	}
}

func addMockEntityV2(t *testing.T, s *storetest.Store, entity *corev2.Entity) {
	entityConfig, entityState := corev3.V2EntityToV3(entity)

	stateReq := storev2.NewResourceRequestFromResource(context.Background(), entityState)
	stateReq.UsePostgres = true

	configReq := storev2.NewResourceRequestFromResource(context.Background(), entityConfig)

	wConfig, err := storev2.WrapResource(entityConfig)
	if err != nil {
		t.Fatal(err)
	}

	wState, err := storev2.WrapResource(entityState)
	if err != nil {
		t.Fatal(err)
	}

	s.On("Get", stateReq).Return(wState, nil)
	s.On("Get", configReq).Return(wConfig, nil)
}

func addFixtureEntity(store *storetest.Store, namespace, name string) {
	entityState := corev3.FixtureEntityState(name)
	entityState.Metadata.Namespace = namespace

	entityConfig := corev3.FixtureEntityConfig(name)
	entityConfig.Metadata.Namespace = namespace

	stateReq := storev2.NewResourceRequestFromResource(context.Background(), entityState)
	stateReq.UsePostgres = true

	configReq := storev2.NewResourceRequestFromResource(context.Background(), entityConfig)

	wConfig, err := storev2.WrapResource(entityConfig)
	if err != nil {
		panic(fmt.Sprintf("couldn't wrap fixture resource, that fixture is probably broken in some way (%v)", err))
	}

	wState, err := storev2.WrapResource(entityState)
	if err != nil {
		panic(fmt.Sprintf("couldn't wrap fixture resource, that fixture is probably broken in some way (%v)", err))
	}

	store.On("Get", mock.MatchedBy(func(req storev2.ResourceRequest) bool {
		return req.Name == configReq.Name &&
			req.Namespace == configReq.Namespace &&
			req.StoreName == configReq.StoreName
	})).Return(wConfig, nil)

	store.On("Get", mock.MatchedBy(func(req storev2.ResourceRequest) bool {
		return req.Name == stateReq.Name &&
			req.Namespace == stateReq.Namespace &&
			req.StoreName == stateReq.StoreName
	})).Return(wState, nil)
}

func entityError(s *storetest.Store, namespace, name string, err error) {
	var nilWrapper storev2.Wrapper

	s.On("Get", mock.MatchedBy(func(req storev2.ResourceRequest) bool {
		return req.Name == name && req.Namespace == namespace
	})).Return(nilWrapper, err)
}

func entityNotFound(s *storetest.Store, namespace, name string) {
	entityError(s, namespace, name, &store.ErrNotFound{})
}

func entityLookupError(s *storetest.Store, namespace, name string) {
	entityError(s, namespace, name, errors.New("some error"))
}

func TestWorkerCount(t *testing.T) {
	// TODO(eric): this is tech debt, better to pass in a config with this
	// property and test after New(). Unfortunately, New() requires a working
	// etcd client and would be too heavy for this test.
	const workers = 10
	daemon := Eventd{
		workerCount: workers,
	}
	if got, want := daemon.Workers(), workers; got != want {
		t.Fatalf("bad workers: got %d, want %d", got, want)
	}
}

func TestEventLog(t *testing.T) {
	log, hook := test.NewNullLogger()
	logger = log.WithField("test", "TestLogger")

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockEntityStore := &storetest.Store{}
	mockStore := &mockstore.MockStore{}
	e := newEventd(mockEntityStore, mockStore, bus, newFakeFactory(&fakeSwitchSet{}))
	// eventd does not panic when started with invalid path
	e.logPath = "/"
	require.NoError(t, e.Start())
	assert.Contains(t, hook.LastEntry().Message, "event log file could not be configured. event logs will not be recorded.")
	require.NoError(t, e.Stop())
}

func TestEventd_handleMessage(t *testing.T) {
	type busFunc func(*mockbus.MockBus)
	type cacheFunc func(*mockcache.MockCache)
	type eventStoreFunc func(*mockstore.MockStore)
	type storeFunc func(*storetest.Store)

	var nilEvent *corev2.Event

	newEntityConfig := func() storev2.Wrapper {
		entity := corev3.FixtureEntityConfig("foo")
		e, err := storev2.WrapResource(entity)
		require.NoError(t, err)
		return e
	}
	newEntityState := func() storev2.Wrapper {
		state := corev3.FixtureEntityState("foo")
		e, err := storev2.WrapResource(state)
		require.NoError(t, err)
		return e
	}

	tests := []struct {
		name           string
		event          corev2.Event
		busFunc        busFunc
		cacheFunc      cacheFunc
		eventStoreFunc eventStoreFunc
		storeFunc      storeFunc
		wantErr        bool
	}{
		{
			name: "metrics events are published without being stored",
			event: corev2.Event{
				Entity:  corev2.FixtureEntity("foo"),
				Metrics: &corev2.Metrics{},
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", messaging.TopicEvent, mock.Anything).Once().Return(nil)
			},
		},
		{
			name: "silenced events are properly identified",
			event: corev2.Event{
				Check:  corev2.FixtureCheck("check-cpu"),
				Entity: corev2.FixtureEntity("foo"),
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", messaging.TopicEvent, mock.Anything).Once().Return(nil)
			},
			cacheFunc: func(c *mockcache.MockCache) {
				c.On("Get", "default").Once().Return(
					[]cache.Value{
						{Resource: corev2.FixtureSilenced("linux:check-cpu")},
					},
				)
			},
			eventStoreFunc: func(store *mockstore.MockStore) {
				store.On("UpdateEvent", mock.AnythingOfType("*v2.Event")).
					Run(func(args mock.Arguments) {
						event := args[0].(*corev2.Event)

						// The event we receive should be identified as silenced and contain
						// the list of silenced entries that apply to it
						if !event.Check.IsSilenced || len(event.Check.Silenced) == 0 {
							t.Fatal("the check should be silenced")
						}
					}).Return(
					corev2.FixtureEvent("foo", "check-cpu"), nilEvent, nil,
				)
			},
			storeFunc: func(store *storetest.Store) {
				store.On("Get", mock.Anything).Once().Return(
					newEntityConfig(), nil,
				)
				store.On("Get", mock.Anything).Once().Return(
					newEntityState(), nil,
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := &mockbus.MockBus{}
			if tt.busFunc != nil {
				tt.busFunc(bus)
			}
			cache := &mockcache.MockCache{}
			if tt.cacheFunc != nil {
				tt.cacheFunc(cache)
			}
			eventStore := &mockstore.MockStore{}
			if tt.eventStoreFunc != nil {
				tt.eventStoreFunc(eventStore)
			}
			store := &storetest.Store{}
			if tt.storeFunc != nil {
				tt.storeFunc(store)
			}
			switches := &mockSwitchSet{}

			e := &Eventd{
				bus:             bus,
				store:           store,
				eventStore:      eventStore,
				livenessFactory: newFakeFactory(switches),
				workerCount:     1,
				wg:              &sync.WaitGroup{},
				Logger:          NoopLogger{},
				silencedCache:   cache,
			}
			if _, err := e.handleMessage(&tt.event); (err != nil) != tt.wantErr {
				t.Errorf("Eventd.handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
