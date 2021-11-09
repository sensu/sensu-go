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
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/handler"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/mocktransport"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var wrappedConfig storev2.Wrapper

func init() {
	var err error
	cfg := corev3.FixtureEntityConfig("testing")
	wrappedConfig, err = storev2.WrapResource(cfg)
	if err != nil {
		panic(err)
	}
}

func TestGoodSessionConfigProto(t *testing.T) {
	conn := new(mocktransport.MockTransport)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start(ctx))

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
	session, err := NewSession(cfg)
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

func TestSession_sender(t *testing.T) {
	type busFunc func(*messaging.WizardBus, *sync.WaitGroup)
	type connFunc func(*mocktransport.MockTransport, *sync.WaitGroup)
	type storeFunc func(*storetest.Store, *sync.WaitGroup)

	tests := []struct {
		name          string
		busFunc       busFunc
		connFunc      connFunc
		storeFunc     storeFunc
		subscriptions []string
	}{
		{
			name: "watch events are propagated to the agents",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				wg.Add(1)
				var once sync.Once
				conn.On("Send", mock.Anything).Run(func(args mock.Arguments) {
					defer once.Do(func() { wg.Done() })
					// Assert the message type to make sure it's an entity update
					msg := args[0].(*transport.Message)
					if msg.Type != transport.MessageTypeEntityConfig {
						t.Fatalf("expected message type %s, got %s", transport.MessageTypeEntityConfig, msg.Type)
					}
				}).Return(nil)
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Closed").Return(true)
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchUpdate,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)
				wg.Wait()
			},
			storeFunc: func(st *storetest.Store, wg *sync.WaitGroup) {
				st.On("Get", mock.Anything).Return(wrappedConfig, nil)
			},
		},
		{
			name: "delete watch event stops the agent session",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				conn.On("SendCloseMessage").Return(nil)
				conn.On("Closed").Return(true)
				conn.On("Close").Return(nil)
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Send", mock.Anything).Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchDelete,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
		},
		{
			name: "unknown watch event are ignored",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				// The Send() method should only be called once, otherwise it means the
				// unknown event also sent something
				conn.On("Send", mock.Anything).Once().Return(transport.ClosedError{})
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Closed").Return(true)
				conn.On("Close").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
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
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
		},
		{
			name: "invalid class entities are reset to the agent class",
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
				entity := corev3.FixtureEntityConfig("testing")
				entity.EntityClass = corev2.EntityProxyClass
				e := store.WatchEventEntityConfig{
					Action: store.WatchUpdate,
					Entity: entity,
				}
				publishWatchEvent(t, bus, e)
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				wg.Add(1)
				s.On("CreateOrUpdate", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					// Close the wait channel once we receive the storev2 request
					wg.Done()
				}).Return(nil)
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				conn.On("Send", mock.Anything).Return(nil)
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Close").Return(nil)
				conn.On("Closed").Return(true)
			},
		},
		{
			name: "the session terminates on send error",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				conn.On("Send", mock.Anything).Return(transport.ConnectionError{Message: "some horrible network outage"})
				conn.On("Closed").Return(true)
				conn.On("Close").Return(nil)
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
			},
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
				if err := bus.Publish(messaging.SubscriptionTopic("default", "testing"), corev2.FixtureCheckRequest("foo")); err != nil {
					t.Fatal(err)
				}
			},
			subscriptions: []string{"testing"},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
		},
		{
			name: "subscriptions are added and check requests are received",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				wg.Add(1)
				var once sync.Once
				conn.On("Send", mock.Anything).Run(func(args mock.Arguments) {
					msg := args[0].(*transport.Message)
					// We expect to receive a message type of both these types
					switch msg.Type {
					case transport.MessageTypeEntityConfig:
						once.Do(func() { wg.Done() })
						t.Logf("received a transport message of type %s", transport.MessageTypeEntityConfig)
					case corev2.CheckRequestType:
						wg.Done()
						t.Logf("received a transport message of type %s", corev2.CheckRequestType)
					}
				}).Return(nil)
				conn.On("Receive").Return(&transport.Message{}, nil)
				conn.On("Closed").Return(false)
				conn.On("Close").Return(nil)
				conn.On("SendCloseMessage").Return(nil)
			},
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchUpdate,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)

				// Wait for the session to subscribe to our new subscription, and then
				// send a check request to this new subscription "linux"
				wg.Wait()
				// TODO(eric): the original tests incorrectly conflate waiting on a waitgroup
				// with session subscription being complete. An additional wait is necessary.
				// These tests are pretty brittle and the session as a whole could probably stand
				// some refactoring.
				time.Sleep(time.Second)
				wg.Add(1)
				if err := bus.Publish(messaging.SubscriptionTopic("default", "linux"), corev2.FixtureCheckRequest("foo")); err != nil {
					t.Fatal(err)
				}
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
			subscriptions: []string{corev2.GetEntitySubscription("testing")},
		},
		{
			name: "subscriptions are removed and check requests are no longer received",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				wg.Add(1)
				var once sync.Once
				conn.On("Send", mock.Anything).Run(func(args mock.Arguments) {
					msg := args[0].(*transport.Message)
					// We only expect to receive an entity config message
					switch msg.Type {
					case transport.MessageTypeEntityConfig:
						t.Logf("received a transport message of type %s", transport.MessageTypeEntityConfig)
						once.Do(func() { wg.Done() })
					case corev2.CheckRequestType:
						// The actual error message might not get reported if we hit this
						// branch, because of a race condition, but the test case will fail
						t.Fatalf("did not expect to receive a message of type %s", corev2.CheckRequestType)
					}
				}).Return(nil)
				conn.On("SendCloseMessage").Return(nil)
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Close").Return(nil)
				conn.On("Closed").Return(false)
			},
			busFunc: func(bus *messaging.WizardBus, wg *sync.WaitGroup) {
				e := store.WatchEventEntityConfig{
					Action: store.WatchUpdate,
					Entity: corev3.FixtureEntityConfig("testing"),
				}
				publishWatchEvent(t, bus, e)

				// Wait for the session to subscribe to our new subscription, and then
				// send a check request to this old subscription "windows"
				wg.Wait()
				if err := bus.Publish(messaging.SubscriptionTopic("default", "windows"), corev2.FixtureCheckRequest("foo")); err != nil {
					t.Fatal(err)
				}
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
			subscriptions: []string{corev2.GetEntitySubscription("testing"), "linux", "windows"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := &sync.WaitGroup{}

			// Mock our transport
			conn := new(mocktransport.MockTransport)
			if tt.connFunc != nil {
				tt.connFunc(conn, wg)
			}

			// Mock our store
			st := &mockstore.MockStore{}
			storev2 := &storetest.Store{}
			if tt.storeFunc != nil {
				tt.storeFunc(storev2, wg)
			}

			// Mock our bus
			bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := bus.Start(ctx); err != nil {
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
			session, err := NewSession(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if err := session.Start(ctx); err != nil {
				t.Fatal(err)
			}

			topic := messaging.EntityConfigTopic(session.cfg.Namespace, session.cfg.AgentName)
			_, err = session.bus.Subscribe(topic, session.cfg.AgentName, session.entityConfig)
			if err != nil {
				t.Fatal(err)
			}

			// Send our watch events over the wizard bus
			if tt.busFunc != nil {
				tt.busFunc(bus, wg)
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				cancel()
				close(done)
			}()

			select {
			case <-done:
				session.Stop()
			case <-time.After(5 * time.Second):
				t.Fatal("session never stopped, we probably never received an entity update over the channel")
			}
		})
	}
}

func TestSession_Start(t *testing.T) {
	type connFunc func(*mocktransport.MockTransport, *sync.WaitGroup)
	type storeFunc func(*storetest.Store, *sync.WaitGroup)

	tests := []struct {
		name      string
		connFunc  connFunc
		storeFunc storeFunc
		wantErr   bool
	}{
		{
			name: "a new entity receives an entity config with EntityNotFound",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Closed").Return(true)
				conn.On("Send", mock.Anything).Run(func(args mock.Arguments) {
					msg := args[0].(*transport.Message)
					var entity corev3.EntityConfig
					if err := UnmarshalJSON(msg.Payload, &entity); err != nil {
						t.Fatal(err)
					}
					if entity.Metadata.Name != corev3.EntityNotFound {
						t.Fatalf("expected entity name %s, got %s", corev3.EntityNotFound, entity.Metadata.Name)
					}
				}).Return(nil)
				conn.On("Close").Return(nil)
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, &store.ErrNotFound{})
			},
		},
		{
			name: "an existing entity receives its stored entity config",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Closed").Return(true)
				conn.On("Send", mock.Anything).Run(func(args mock.Arguments) {
					msg := args[0].(*transport.Message)
					var entity corev3.EntityConfig
					if err := UnmarshalJSON(msg.Payload, &entity); err != nil {
						t.Fatal(err)
					}
					if entity.Metadata.Name != "testing" {
						t.Fatalf("expected entity name %s, got %s", "testing", entity.Metadata.Name)
					}
				}).Return(nil)
				conn.On("Close").Return(nil)
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(wrappedConfig, nil)
			},
		},
		{
			name: "store err is handled",
			connFunc: func(conn *mocktransport.MockTransport, wg *sync.WaitGroup) {
				conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
				conn.On("Closed").Return(true)
				conn.On("Close").Return(nil)
			},
			storeFunc: func(s *storetest.Store, wg *sync.WaitGroup) {
				s.On("Get", mock.Anything).Return(&wrap.Wrapper{}, errors.New("fatal error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := &sync.WaitGroup{}

			// Mock our transport
			conn := new(mocktransport.MockTransport)
			if tt.connFunc != nil {
				tt.connFunc(conn, wg)
			}

			// Mock our store
			st := &mockstore.MockStore{}
			storev2 := &storetest.Store{}
			if tt.storeFunc != nil {
				tt.storeFunc(storev2, wg)
			}

			// Mock our bus
			bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := bus.Start(ctx); err != nil {
				t.Fatal(err)
			}

			cfg := SessionConfig{
				AgentName: "testing",
				Namespace: "default",
				Conn:      conn,
				Bus:       bus,
				Store:     st,
				Storev2:   storev2,
				Unmarshal: UnmarshalJSON,
				Marshal:   MarshalJSON,
			}
			session, err := NewSession(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if err := session.Start(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Session.Start() error = %v, wantErr %v", err, tt.wantErr)
			}

			done := make(chan struct{})
			go func() {
				wg.Wait()
				cancel()
				close(done)
			}()

			select {
			case <-done:
				session.Stop()
			case <-time.After(5 * time.Second):
				t.Fatal("session never stopped, we probably never received an entity update over the channel")
			}

			// Make sure the check subscriptions were all cancelled
			session.mu.Lock()
			defer session.mu.Unlock()
			if len(session.subscriptionsMap) > 0 {
				t.Fatalf("expected all check subsriptions to be cancelled, found %#v\n", session.subscriptionsMap)
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

func Test_removeEmptySubscriptions(t *testing.T) {
	tests := []struct {
		name          string
		subscriptions []string
		want          []string
	}{
		{
			name:          "no empty subscriptions",
			subscriptions: []string{"foo", "bar"},
			want:          []string{"foo", "bar"},
		},
		{
			name:          "leading empty subscriptions",
			subscriptions: []string{"", "foo", "bar"},
			want:          []string{"foo", "bar"},
		},
		{
			name:          "middle empty subscriptions",
			subscriptions: []string{"foo", "", "bar"},
			want:          []string{"foo", "bar"},
		},
		{
			name:          "trailing empty subscriptions",
			subscriptions: []string{"foo", "bar", ""},
			want:          []string{"foo", "bar"},
		},
		{
			name:          "multiple empty subscriptions",
			subscriptions: []string{"", "foo", "bar", ""},
			want:          []string{"foo", "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeEmptySubscriptions(tt.subscriptions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeEmptySubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSession_receiver(t *testing.T) {
	type connFunc func(*mocktransport.MockTransport, context.CancelFunc)

	tests := []struct {
		name     string
		connFunc connFunc
	}{
		{
			name: "incoming messages are handled",
			connFunc: func(conn *mocktransport.MockTransport, cancel context.CancelFunc) {
				conn.On("Receive").Once().Return(&transport.Message{}, nil)
				conn.On("Receive").Once().Run(func(args mock.Arguments) {
					cancel()
				}).Return(&transport.Message{}, nil)
			},
		},
		{
			name: "random errors are handled",
			connFunc: func(conn *mocktransport.MockTransport, cancel context.CancelFunc) {
				conn.On("Receive").Once().Return(&transport.Message{}, errors.New("error"))
				conn.On("Receive").Once().Run(func(args mock.Arguments) {
					cancel()
				}).Return(&transport.Message{}, nil)
			},
		},
		{
			name: "transport errors are handled",
			connFunc: func(conn *mocktransport.MockTransport, cancel context.CancelFunc) {
				conn.On("Receive").Once().Return(&transport.Message{}, transport.ConnectionError{})
				conn.On("Receive").Once().Run(func(args mock.Arguments) {
					cancel()
				}).Return(&transport.Message{}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			conn := new(mocktransport.MockTransport)
			if tt.connFunc != nil {
				tt.connFunc(conn, cancel)
			}

			s := &Session{
				cfg: SessionConfig{
					WriteTimeout: 5,
				},
				conn: conn,
				wg:   &sync.WaitGroup{},
			}
			s.wg.Add(1)
			s.handler = handler.NewMessageHandler()
			go s.receiver(ctx)

			s.wg.Wait()
		})
	}
}
