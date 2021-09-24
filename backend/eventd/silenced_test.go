package eventd

import (
	"context"
	"sync"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/cache"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetSilenced(t *testing.T) {
	testCases := []struct {
		name            string
		event           *corev2.Event
		silencedEntries []corev2.Resource
		expectedEntries []string
	}{
		{
			name:  "Sets the silenced attribute of an event",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			silencedEntries: []corev2.Resource{
				corev2.FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:check_cpu"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
			c := cache.NewFromResources(tc.silencedEntries, false)

			getSilenced(ctx, tc.event, c)
			assert.Equal(t, tc.expectedEntries, tc.event.Check.Silenced)
		})
	}
}

func TestSilencedBy(t *testing.T) {
	testCases := []struct {
		name            string
		event           *corev2.Event
		entries         []*corev2.Silenced
		expectedEntries []string
	}{
		{
			name:            "no entries",
			event:           corev2.FixtureEvent("foo", "check_cpu"),
			expectedEntries: []string{},
		},
		{
			name:  "not silenced",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:check_mem"),
				corev2.FixtureSilenced("entity:bar:*"),
				corev2.FixtureSilenced("foo:check_cpu"),
				corev2.FixtureSilenced("foo:*"),
				corev2.FixtureSilenced("*:check_mem"),
			},
			expectedEntries: []string{},
		},
		{
			name:  "silenced by check",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("*:check_cpu"),
			},
			expectedEntries: []string{"*:check_cpu"},
		},
		{
			name:  "silenced by entity subscription",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []string{"entity:foo:*"},
		},
		{
			name:  "silenced by entity's check subscription",
			event: corev2.FixtureEvent("foo", "check_cpu"),
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:check_cpu"},
		},
		{
			name:  "silenced by check subscription",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("linux:*"),
			},
			expectedEntries: []string{"linux:*"},
		},
		{
			name:  "silenced by subscription with check",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"linux:check_cpu"},
		},
		{
			name:  "silenced by multiple entries",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:*"),
				corev2.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"entity:foo:*", "linux:check_cpu"},
		},
		{
			name:  "silenced by duplicated entries",
			event: corev2.FixtureEvent("foo", "check_cpu"), // has a linux subscription
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("entity:foo:*"),
				corev2.FixtureSilenced("entity:foo:*"),
			},
			expectedEntries: []string{"entity:foo:*"},
		},
		{
			name: "not silenced, silenced & client don't have a common subscription",
			event: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("windows:check_cpu"),
			},
			expectedEntries: []string{},
		},
		{
			name: "silenced, silenced & client do have a common subscription",
			event: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name: "check_cpu",
					},
					Subscriptions: []string{"linux", "windows"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name: "foo",
					},
					Subscriptions: []string{"linux"},
				},
			},
			entries: []*corev2.Silenced{
				corev2.FixtureSilenced("linux:check_cpu"),
			},
			expectedEntries: []string{"linux:check_cpu"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := silencedBy(tc.event, tc.entries)
			assert.Equal(t, tc.expectedEntries, result)
		})
	}
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(namespace string) []cache.Value {
	args := m.Called(namespace)
	return args.Get(0).([]cache.Value)
}

func TestEventd_handleMessage(t *testing.T) {
	type busFunc func(*mockbus.MockBus)
	type cacheFunc func(*mockCache)
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
			cacheFunc: func(c *mockCache) {
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
			cache := &mockCache{}
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
			if err := e.handleMessage(&tt.event); (err != nil) != tt.wantErr {
				t.Errorf("Eventd.handleMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
