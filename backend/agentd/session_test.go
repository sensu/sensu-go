package agentd

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/mocktransport"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGoodSessionConfigProto(t *testing.T) {
	conn := new(mocktransport.MockTransport)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	st := &mockstore.MockStore{}
	st.On(
		"GetNamespace",
		mock.Anything,
		"acme",
	).Return(&corev2.Namespace{}, nil)

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{"testing"},
		Conn:          conn,
		Bus:           bus,
		Store:         st,
		Unmarshal:     proto.Unmarshal,
		Marshal:       proto.Marshal,
	}
	session, err := NewSession(context.Background(), cfg)
	assert.NotNil(t, session)
	assert.NoError(t, err)
}

func TestMakeEntitySwitchBurialEvent(t *testing.T) {
	cfg := SessionConfig{
		Namespace:     "default",
		AgentName:     "entity",
		Subscriptions: []string{"default"},
	}
	event := makeEntitySwitchBurialEvent(cfg)
	if err := event.Validate(); err != nil {
		t.Fatal(err)
	}
	if err := event.Entity.Validate(); err != nil {
		t.Fatal(err)
	}
	if got, want := event.Timestamp, int64(deletedEventSentinel); got != want {
		t.Errorf("bad timestamp: got %d, want %d", got, want)
	}
}

func TestSession(t *testing.T) {
	type busFunc func(*messaging.WizardBus)
	type connFunc func(*mocktransport.MockTransport, chan struct{})
	type storeFunc func(*storetest.Store, chan struct{})

	tests := []struct {
		name          string
		busFunc       busFunc
		connFunc      connFunc
		storeFunc     storeFunc
		subscriptions []string
	}{
		{
			name: "watch events are propagated to the agents",
			connFunc: func(conn *mocktransport.MockTransport, wait chan struct{}) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Send", mock.Anything).Run(func(args mock.Arguments) {
					// Assert the message type to make sure it's an entity update
					msg := args[0].(*transport.Message)
					if msg.Type != transport.MessageTypeEntityConfig {
						t.Fatalf("expected message type %s, got %s", transport.MessageTypeEntityConfig, msg.Type)
					}

					// Close our wait channel once we asserted the message
					close(wait)
				}).Return(nil)
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchUpdate,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)
			},
		},
		{
			name: "delete watch event stops the agent session",
			connFunc: func(conn *mocktransport.MockTransport, wait chan struct{}) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchDelete,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)
			},
		},
		{
			name: "unknown watch event are ignored",
			connFunc: func(conn *mocktransport.MockTransport, wait chan struct{}) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				// The Send() method should only be called once, otherwise it means the
				// unknown event also sent something
				conn.On("Send", mock.Anything).Once().Return(transport.ClosedError{})
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchUnknown,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)

				// publish a second valid events, which will trigger the Send() method
				// of our transport, which will mock a closed connection that should
				// only be called once
				e = store.WatchEventEntityConfig{
					Action: store.WatchCreate,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)
			},
		},
		{
			name: "invalid class entities are reset to the agent class",
			connFunc: func(conn *mocktransport.MockTransport, wait chan struct{}) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus) {
				entity := corev3.FixtureEntityConfig("testing")
				entity.EntityClass = corev2.EntityProxyClass
				e := store.WatchEventEntityConfig{
					Action: store.WatchUpdate,
					Entity: entity,
				}
				publishWatchEvent(t, bus, e)
			},
			storeFunc: func(store *storetest.Store, wait chan struct{}) {
				fmt.Println("storeFunc!")
				store.On("CreateOrUpdate", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					// Close the wait channel once we receive the storev2 request
					close(wait)
				}).Return(nil)
			},
		},
		{
			name: "the session terminates on send error",
			connFunc: func(conn *mocktransport.MockTransport, wait chan struct{}) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Send", mock.Anything).Return(transport.ConnectionError{Message: "some horrible network outage"})
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus) {
				if err := bus.Publish(messaging.SubscriptionTopic("default", "testing"), corev2.FixtureCheckRequest("foo")); err != nil {
					t.Fatal(err)
				}
			},
			subscriptions: []string{"testing"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait := make(chan struct{})

			// Mock our transport
			conn := new(mocktransport.MockTransport)
			if tt.connFunc != nil {
				tt.connFunc(conn, wait)
			}

			// Mock our store
			st := &mockstore.MockStore{}
			storev2 := &storetest.Store{}
			if tt.storeFunc != nil {
				tt.storeFunc(storev2, wait)
			}

			// Mock our bus
			bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
			if err != nil {
				t.Fatal(err)
			}
			if err := bus.Start(); err != nil {
				t.Fatal(err)
			}

			cfg := SessionConfig{
				AgentName:     "testing",
				Namespace:     "default",
				Subscriptions: tt.subscriptions,
				Conn:          conn,
				Bus:           bus,
				Store:         st,
				Storev2:       storev2,
				Unmarshal:     UnmarshalJSON,
				Marshal:       MarshalJSON,
			}
			session, err := NewSession(context.Background(), cfg)
			if err != nil {
				t.Fatal(err)
			}
			if err := session.Start(); err != nil {
				t.Fatal(err)
			}

			// Send our watch events over the wizard bus
			if tt.busFunc != nil {
				tt.busFunc(bus)
			}

			select {
			case <-session.ctx.Done():
			case <-wait:
				session.Stop()
			case <-time.After(5 * time.Second):
				t.Fatal("session never stopped, we probably never received an entity update over the channel")
			}
		})
	}
}

func publishWatchEvent(t *testing.T, bus *messaging.WizardBus, event store.WatchEventEntityConfig) {
	t.Helper()

	if err := bus.Publish(messaging.EntityConfigTopic(
		event.Entity.Metadata.Namespace, event.Entity.Metadata.Name,
	), &event); err != nil {
		t.Fatal(err)
	}
}

func TestSession_subscribe(t *testing.T) {
	type busFunc func(*mockbus.MockBus)

	fooTopic := fmt.Sprintf("%s:%s:%s", messaging.TopicSubscriptions, "default", "foo")

	tests := []struct {
		name             string
		subscriptions    []string
		busFunc          busFunc
		subscriptionsMap map[string]subscription
		want             map[string]subscription
		wantErr          bool
	}{
		{
			name:          "empty subscriptions are ignored",
			subscriptions: []string{""},
		},
		{
			name:          "already subscribed subscriptions are ignored",
			subscriptions: []string{"foo"},
			subscriptionsMap: map[string]subscription{
				fooTopic: &messaging.Subscription{},
			},
			want: map[string]subscription{
				fooTopic: &messaging.Subscription{},
			},
		},
		{
			name:          "subscriptions are successfully performed",
			subscriptions: []string{"foo"},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Subscribe", "sensu:check:default:foo", mock.Anything, mock.Anything).
					Return(messaging.Subscription{}, nil)
			},
			subscriptionsMap: map[string]subscription{},
			want: map[string]subscription{
				fooTopic: &messaging.Subscription{},
			},
		},
		{
			name:          "bus errors are handled",
			subscriptions: []string{"bar"},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Subscribe", "sensu:check:default:bar", mock.Anything, mock.Anything).
					Return(messaging.Subscription{}, errors.New("error"))
			},
			subscriptionsMap: map[string]subscription{},
			want:             map[string]subscription{},
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := &mockbus.MockBus{}
			if tt.busFunc != nil {
				tt.busFunc(bus)
			}

			s := &Session{
				cfg: SessionConfig{
					AgentName:     "foo",
					Namespace:     "default",
					Subscriptions: tt.subscriptions,
				},
				bus:              bus,
				mu:               sync.Mutex{},
				subscriptionsMap: tt.subscriptionsMap,
			}
			if err := s.subscribe(tt.subscriptions); (err != nil) != tt.wantErr {
				t.Errorf("Session.subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(s.subscriptionsMap, tt.want) {
				t.Errorf("Session.subscribe() subscriptionsMap = %v, want %v", s.subscriptionsMap, tt.want)
			}
		})
	}
}

// mockSubscription mocks a messaging.Subscription
type mockSubscription struct {
	mock.Mock
}

// Cancel ...
func (m *mockSubscription) Cancel() error {
	args := m.Called()
	return args.Error(0)
}

func TestSession_unsubscribe(t *testing.T) {
	type subscriptionFunc func(*mockSubscription)

	mockedSubscription := &mockSubscription{}

	fooTopic := fmt.Sprintf("%s:%s:%s", messaging.TopicSubscriptions, "default", "foo")

	tests := []struct {
		name             string
		subscriptions    []string
		subscriptionFunc subscriptionFunc
		subscriptionsMap map[string]subscription
		want             map[string]subscription
	}{
		{
			name:          "subscriptions can be successfully unsubscribed from",
			subscriptions: []string{"foo"},
			subscriptionFunc: func(subscription *mockSubscription) {
				subscription.On("Cancel").Return(nil)
			},
			subscriptionsMap: map[string]subscription{
				fooTopic: &mockSubscription{},
			},
			want: map[string]subscription{},
		},
		{
			name:             "subscriptions the session is not subscribed to already are ignored",
			subscriptions:    []string{"foo"},
			subscriptionsMap: map[string]subscription{},
			want:             map[string]subscription{},
		},
		{
			name:          "errors from subscriptions are handled",
			subscriptions: []string{"foo"},
			subscriptionFunc: func(subscription *mockSubscription) {
				subscription.On("Cancel").Return(errors.New("error"))
			},
			subscriptionsMap: map[string]subscription{
				fooTopic: mockedSubscription,
			},
			want: map[string]subscription{
				fooTopic: mockedSubscription,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				cfg: SessionConfig{
					AgentName:     "foo",
					Namespace:     "default",
					Subscriptions: tt.subscriptions,
				},
				mu:               sync.Mutex{},
				subscriptionsMap: tt.subscriptionsMap,
			}

			if tt.subscriptionFunc != nil {
				tt.subscriptionFunc(s.subscriptionsMap[fooTopic].(*mockSubscription))
			}

			s.unsubscribe(tt.subscriptions)

			if !reflect.DeepEqual(s.subscriptionsMap, tt.want) {
				t.Errorf("Session.unsubscribe() subscriptionsMap = %v, want %v", s.subscriptionsMap, tt.want)
			}
		})
	}
}

func Test_diff(t *testing.T) {
	tests := []struct {
		name        string
		old         []string
		new         []string
		wantAdded   []string
		wantRemoved []string
	}{
		{
			name:        "simple removed and added elements",
			old:         []string{"a", "b", "c"},
			new:         []string{"b", "c", "d"},
			wantAdded:   []string{"d"},
			wantRemoved: []string{"a"},
		},
		{
			name:        "simple removed and added elements but reversed",
			old:         []string{"b", "c", "d"},
			new:         []string{"a", "b", "c"},
			wantAdded:   []string{"a"},
			wantRemoved: []string{"d"},
		},
		{
			name:      "duplicated elements are detected",
			old:       []string{"a", "b", "c"},
			new:       []string{"a", "a", "b", "c"},
			wantAdded: []string{"a"},
		},
		{
			name:        "completely different slices",
			old:         []string{"a", "b", "c"},
			new:         []string{"d", "e", "f"},
			wantAdded:   []string{"d", "e", "f"},
			wantRemoved: []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdded, gotRemoved := diff(tt.old, tt.new)
			if !reflect.DeepEqual(gotAdded, tt.wantAdded) {
				t.Errorf("diff() added = %#v, want %#v", gotAdded, tt.wantAdded)
			}
			if !reflect.DeepEqual(gotRemoved, tt.wantRemoved) {
				t.Errorf("diff() removed = %#v, want %#v", gotRemoved, tt.wantRemoved)
			}
		})
	}
}

func Test_sortSubscriptions(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []string
		want          []string
	}{
		{
			name:          "unsorted subscriptions are sorted",
			subscriptions: []string{"b", "a", "c"},
			want:          []string{"a", "b", "c"},
		},
		{
			name:          "already sorted subscriptions are immediately returned",
			subscriptions: []string{"a", "b", "c"},
			want:          []string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortSubscriptions(tt.subscriptions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortSubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
