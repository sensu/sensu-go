package eventd

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/liveness"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

func newFakeFactory(f liveness.Interface) liveness.Factory {
	return func(name string, dead, alive liveness.EventFunc, logger logrus.FieldLogger) liveness.Interface {
		return f
	}
}

func newEventd(store store.Store, bus messaging.MessageBus, livenessFactory liveness.Factory) *Eventd {
	ctx, cancel := context.WithCancel(context.Background())
	return &Eventd{
		ctx:             ctx,
		cancel:          cancel,
		store:           store,
		eventStore:      store,
		bus:             bus,
		livenessFactory: livenessFactory,
		errChan:         make(chan error, 1),
		shutdownChan:    make(chan struct{}, 1),
		eventChan:       make(chan interface{}, 100),
		wg:              &sync.WaitGroup{},
		mu:              &sync.Mutex{},
		Logger:          &RawLogger{},
		workerCount:     5,
		storeTimeout:    time.Minute,
		silencedCache:   &cache.Resource{},
	}
}

func TestEventHandling(t *testing.T) {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	mockStore := &mockstore.MockStore{}
	e := newEventd(mockStore, bus, newFakeFactory(&fakeSwitchSet{}))

	require.NoError(t, e.Start())

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, nil))

	badEvent := &corev2.Event{}
	badEvent.Check = &corev2.Check{}
	badEvent.Entity = &corev2.Entity{}
	badEvent.Timestamp = time.Now().Unix()

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, badEvent))

	event := corev2.FixtureEvent("entity", "check")

	mockStore.On("GetEntityByName", mock.Anything, "entity").
		Return(event.Entity, nil)

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

	mockStore := &mockstore.MockStore{}
	e := newEventd(mockStore, bus, newFakeFactory(&fakeSwitchSet{}))

	require.NoError(t, e.Start())

	require.NoError(t, bus.Publish(messaging.TopicEventRaw, nil))

	event := corev2.FixtureEvent("entity", "check")
	event.Check.Ttl = 90

	mockStore.On("GetEntityByName", mock.Anything, "entity").
		Return(event.Entity, nil)

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
			store := &mockstore.MockStore{}
			switches := &mockSwitchSet{}
			if tt.switchesFunc != nil {
				tt.switchesFunc(switches)
			}

			mockEvent := corev2.FixtureEvent("entity", "mock")

			e := &Eventd{
				store:           store,
				eventStore:      store,
				livenessFactory: newFakeFactory(switches),
				workerCount:     1,
				wg:              &sync.WaitGroup{},
				Logger:          &RawLogger{},
				silencedCache:   &cache.Resource{},
			}
			var err error
			e.bus, err = messaging.NewWizardBus(messaging.WizardBusConfig{})
			require.NoError(t, err)
			require.NoError(t, e.bus.Start())

			store.On("GetEntityByName", mock.Anything, "entity").
				Return(mockEvent.Entity, nil)
			store.On("GetEventByEntityCheck", mock.Anything, "entity", "check").
				Return(tt.previousEvent, tt.previousEventErr)
			store.On("GetSilencedEntriesBySubscription", mock.Anything, mock.Anything).
				Return([]*corev2.Silenced{}, nil)
			store.On("GetSilencedEntriesByCheckName", mock.Anything, mock.Anything).
				Return([]*corev2.Silenced{}, nil)
			store.On("UpdateEvent", mock.Anything, mock.Anything).Return(tt.msg, mockEvent, nil)

			if err := e.handleMessage(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("Eventd.handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuryConditions(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		store func(*mockstore.MockStore)
		bury  bool
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
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return((*corev2.Entity)(nil), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: true,
		},
		{
			name: "do not bury switch on entity lookup error",
			key:  "default/foo/bar",
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return((*corev2.Entity)(nil), errors.New("error"))
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: false,
		},
		{
			name: "bury switch on nil event",
			key:  "default/foo/bar",
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return(corev2.FixtureEntity("bar"), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return((*corev2.Event)(nil), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: true,
		},
		{
			name: "do not bury switch on event lookup error",
			key:  "default/foo/bar",
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return(corev2.FixtureEntity("bar"), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return((*corev2.Event)(nil), errors.New("!"))
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: false,
		},
		{
			name: "do not bury switch otherwise",
			key:  "default/foo/bar",
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return(corev2.FixtureEntity("bar"), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "foo").Return(corev2.FixtureEvent("bar", "foo"), nil)
				store.On("GetEventByEntityCheck", mock.Anything, "bar", "keepalive").Return(corev2.FixtureEvent("bar", "keepalive"), nil)
			},
			bury: false,
		},
		{
			name: "bury when failing keepalive exists",
			key:  "default/foo/bar",
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return(corev2.FixtureEntity("bar"), nil)
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
			store: func(store *mockstore.MockStore) {
				store.On("GetEntityByName", mock.Anything, "bar").Return(corev2.FixtureEntity("bar"), nil)
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
			store := new(mockstore.MockStore)
			if test.store != nil {
				test.store(store)
			}
			eventd := &Eventd{store: store, eventStore: store, ctx: context.Background()}
			if got, want := eventd.dead(test.key, liveness.Alive, false), test.bury; got != want {
				t.Fatalf("bad bury result: got %v, want %v", got, want)
			}
		})
	}
}
