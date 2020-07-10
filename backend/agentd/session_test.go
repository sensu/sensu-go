package agentd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

type testTransport struct {
	sendCh  chan *transport.Message
	closed  bool
	sendErr error
	recvErr error
}

func (t *testTransport) Closed() bool {
	return t.closed
}

func (t *testTransport) Close() error {
	t.closed = true
	return nil
}

func (t *testTransport) Heartbeat(ctx context.Context, interval, timeout int) {}

func (t *testTransport) Reconnect(wsServerURL string, tlsOpts *corev2.TLSOptions, requestHeader http.Header) error {
	return nil
}

func (t *testTransport) Send(msg *transport.Message) error {
	if t.sendErr != nil {
		return t.sendErr
	}
	t.sendCh <- msg
	return nil
}

func (t *testTransport) Receive() (*transport.Message, error) {
	if t.recvErr != nil {
		return nil, t.recvErr
	}
	return <-t.sendCh, nil
}

func TestGoodSessionConfig(t *testing.T) {
	conn := &testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

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
		Unmarshal:     UnmarshalJSON,
		Marshal:       MarshalJSON,
	}
	session, err := NewSession(context.Background(), cfg)
	assert.NotNil(t, session)
	assert.NoError(t, err)
}

func TestGoodSessionConfigProto(t *testing.T) {
	conn := &testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

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

func TestSessionTerminateOnSendError(t *testing.T) {
	conn := new(mocktransport.MockTransport)
	event := corev2.FixtureEvent("acme", "testing")
	b, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	tm := &transport.Message{
		Payload: b,
		Type:    transport.MessageTypeEvent,
	}

	conn.On("Receive").After(100*time.Millisecond).Return(tm, nil)
	conn.On("Send", mock.Anything).Return(transport.ConnectionError{Message: "some horrible network outage"})
	conn.On("Close").Return(nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}

	st := &mockstore.MockStore{}
	st.On("GetNamespace", mock.Anything, "acme").Return(&corev2.Namespace{}, nil)
	st.On("GetEntityByName", mock.Anything, "acme").Return(event.Entity, nil)

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{"testing"},
		Conn:          conn,
		Bus:           bus,
		Store:         st,
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
	if err := bus.Publish(messaging.SubscriptionTopic("acme", "testing"), corev2.FixtureCheckRequest("foo")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-session.ctx.Done():
	case <-time.After(time.Second * 5):
		t.Fatal("broken session never stopped")
	}
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

func TestSessionEntityUpdate(t *testing.T) {
	wait := make(chan struct{})

	conn := new(mocktransport.MockTransport)
	// Mock the Receive method by blocking it for 100ms and returns an empty
	// message so it doesn't block our test for too long
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

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}

	st := &mockstore.MockStore{}

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{""},
		Conn:          conn,
		Bus:           bus,
		Store:         st,
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

	// Send an entity config to mock an update to etcd
	watchEvent := store.WatchEventEntityConfig{
		Action: store.WatchUpdate,
		Entity: corev3.FixtureEntityConfig("testing"),
	}
	if err := bus.Publish(messaging.EntityConfigTopic("acme", "testing"), &watchEvent); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		session.Stop()
	case <-time.After(5 * time.Second):
		t.Fatal("session never stopped, we probably never received an entity update over the channel")
	}
}

func TestSessionEntityWatchDeleteAndUnknown(t *testing.T) {
	conn := new(mocktransport.MockTransport)
	// Mock the Receive method by blocking it for 100ms and returns an empty
	// message so it doesn't block our test for too long
	conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
	conn.On("Close").Return(nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}

	st := &mockstore.MockStore{}

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{""},
		Conn:          conn,
		Bus:           bus,
		Store:         st,
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

	// Mock an unknown watch event, which should only log an entry and continue
	// with the next event
	watchEvent := store.WatchEventEntityConfig{
		Action: store.WatchUnknown,
		Entity: corev3.FixtureEntityConfig("testing"),
	}
	if err := bus.Publish(messaging.EntityConfigTopic("acme", "testing"), &watchEvent); err != nil {
		t.Fatal(err)
	}

	// Mock a delete watch event, which should force the session to close itself
	// so the agent can attempt to reconnect in order to register itself again
	watchEvent2 := store.WatchEventEntityConfig{
		Action: store.WatchDelete,
		Entity: corev3.FixtureEntityConfig("testing"),
	}
	if err := bus.Publish(messaging.EntityConfigTopic("acme", "testing"), &watchEvent2); err != nil {
		t.Fatal(err)
	}

	// The session should close itself upon a WatchDelete
	select {
	case <-session.ctx.Done():
	case <-time.After(time.Second * 5):
		t.Fatal("broken session never stopped")
	}
}

func TestSessionInvalidEntityClassUpdate(t *testing.T) {
	wait := make(chan struct{})

	conn := new(mocktransport.MockTransport)
	// Mock the Receive method by blocking it for 100ms and returns an empty
	// message so it doesn't block our test for too long
	conn.On("Receive").After(100*time.Millisecond).Return(&transport.Message{}, nil)
	conn.On("Close").Return(nil)

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}

	storev2 := &storetest.Store{}
	storev2.On("CreateOrUpdate", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// Close the wait channel once we receive the storev2 request
		close(wait)
	}).Return(nil)

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{""},
		Conn:          conn,
		Bus:           bus,
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

	// Mock an entity update from the entity watcher, and update the entity class
	// to simulate a misconfigured agent entity as a proxy entity
	entity := corev3.FixtureEntityConfig("testing")
	entity.EntityClass = corev2.EntityProxyClass
	watchEvent := store.WatchEventEntityConfig{
		Action: store.WatchUpdate,
		Entity: entity,
	}
	if err := bus.Publish(messaging.EntityConfigTopic("acme", "testing"), &watchEvent); err != nil {
		t.Fatal(err)
	}

	select {
	case <-wait:
		session.Stop()
	case <-time.After(5 * time.Second):
		t.Fatal("session never stopped, we probably never received an entity update over the channel")
	}
}

func TestSession_subscribe(t *testing.T) {
	type busFunc func(*mockbus.MockBus)

	fooTopic := fmt.Sprintf("%s:%s:%s", messaging.TopicSubscriptions, "default", "foo")

	tests := []struct {
		name             string
		subscriptions    []string
		busFunc          busFunc
		subscriptionsMap map[string]messaging.Subscription
		want             map[string]messaging.Subscription
		wantErr          bool
	}{
		{
			name:             "empty subscriptions are ignored",
			subscriptions:    []string{""},
			subscriptionsMap: map[string]messaging.Subscription{},
			want:             map[string]messaging.Subscription{},
		},
		{
			name:          "already subscribed subscriptions are ignored",
			subscriptions: []string{"foo"},
			subscriptionsMap: map[string]messaging.Subscription{
				fooTopic: {},
			},
			want: map[string]messaging.Subscription{
				fooTopic: {},
			},
		},
		{
			name:          "subscriptions are successfully performed",
			subscriptions: []string{"foo"},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Subscribe", "sensu:check:default:foo", mock.Anything, mock.Anything).
					Return(messaging.Subscription{}, nil)
			},
			subscriptionsMap: map[string]messaging.Subscription{},
			want: map[string]messaging.Subscription{
				fooTopic: {},
			},
		},
		{
			name:          "bus errors are handled",
			subscriptions: []string{"bar"},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Subscribe", "sensu:check:default:bar", mock.Anything, mock.Anything).
					Return(messaging.Subscription{}, errors.New("error"))
			},
			subscriptionsMap: map[string]messaging.Subscription{},
			want:             map[string]messaging.Subscription{},
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
