package agentd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
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
	}
	session, err := NewSession(context.Background(), cfg, conn, bus, st, UnmarshalJSON, MarshalJSON)
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
	}
	session, err := NewSession(context.Background(), cfg, conn, bus, st, proto.Unmarshal, proto.Marshal)
	assert.NotNil(t, session)
	assert.NoError(t, err)
}

func TestBadSessionConfig(t *testing.T) {
	conn := &testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	st := &mockstore.MockStore{}
	st.On(
		"UpdateEntity",
		mock.Anything,
		mock.Anything,
	).Return(nil)
	st.On(
		"GetNamespace",
		mock.Anything,
		mock.AnythingOfType("string"),
	).Return(&corev2.Namespace{}, fmt.Errorf("error"))

	cfg := SessionConfig{
		Subscriptions: []string{"testing"},
	}
	session, err := NewSession(context.Background(), cfg, conn, bus, st, UnmarshalJSON, MarshalJSON)
	assert.Nil(t, session)
	assert.Error(t, err)
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
	}
	session, err := NewSession(context.Background(), cfg, conn, bus, st, UnmarshalJSON, MarshalJSON)
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

type channelSubscriber struct {
	ch chan interface{}
}

func (c channelSubscriber) Receiver() chan<- interface{} {
	return c.ch
}

// func TestSession_handleEvent(t *testing.T) {
// 	type busFunc func(*testing.T, *mockbus.MockBus)
// 	type storeFunc func(*mockstore.MockStore)

// 	var nilEntity *types.Entity
// 	event := corev2.FixtureEvent("foo", "check-cpu")
// 	proxyEvent := corev2.FixtureEvent("bar", "check-cpu")
// 	proxyEvent.Check.ProxyEntityName = "baz"

// 	tests := []struct {
// 		name       string
// 		event      *corev2.Event
// 		busFunc    busFunc
// 		storeFunc  storeFunc
// 		wantEntity string
// 		wantErr    bool
// 	}{
// 		{
// 			name:  "event with existing entity gets published",
// 			event: event,
// 			busFunc: func(t *testing.T, bus *mockbus.MockBus) {
// 				bus.On("Publish", messaging.TopicEventRaw, mock.AnythingOfType("*v2.Event")).
// 					Return(nil)
// 			},
// 			storeFunc: func(store *mockstore.MockStore) {
// 				store.On("GetEntityByName", mock.Anything, "foo").
// 					Return(event.Entity, nil)
// 			},
// 		},
// 		{
// 			name:  "event with missing entity gets published",
// 			event: event,
// 			busFunc: func(t *testing.T, bus *mockbus.MockBus) {
// 				bus.On("Publish", messaging.TopicEventRaw, mock.AnythingOfType("*v2.Event")).
// 					Return(nil)
// 			},
// 			storeFunc: func(store *mockstore.MockStore) {
// 				store.On("GetEntityByName", mock.Anything, "foo").
// 					Return(nilEntity, nil)
// 				store.On("UpdateEntity", mock.Anything, mock.AnythingOfType("*v2.Entity")).
// 					Return(nil)
// 			},
// 		},
// 		{
// 			name:  "event with existing proxy entity gets published",
// 			event: proxyEvent,
// 			busFunc: func(t *testing.T, bus *mockbus.MockBus) {
// 				bus.On("Publish", messaging.TopicEventRaw, mock.AnythingOfType("*v2.Event")).
// 					Return(nil)
// 			},
// 			storeFunc: func(store *mockstore.MockStore) {
// 				store.On("GetEntityByName", mock.Anything, "baz").
// 					Return(proxyEvent.Entity, nil)
// 			},
// 		},
// 		{
// 			name:  "event with missing proxy entity gets published",
// 			event: proxyEvent,
// 			busFunc: func(t *testing.T, bus *mockbus.MockBus) {
// 				bus.On("Publish", messaging.TopicEventRaw, mock.AnythingOfType("*v2.Event")).
// 					Return(nil)
// 			},
// 			storeFunc: func(store *mockstore.MockStore) {
// 				store.On("GetEntityByName", mock.Anything, "baz").
// 					Return(nilEntity, nil)
// 				store.On("UpdateEntity", mock.Anything, mock.AnythingOfType("*v2.Entity")).
// 					Return(nil)
// 			},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			store := &mockstore.MockStore{}
// 			if tt.storeFunc != nil {
// 				tt.storeFunc(store)
// 			}
// 			defer store.AssertExpectations(t)

// 			bus := &mockbus.MockBus{}
// 			if tt.busFunc != nil {
// 				tt.busFunc(t, bus)
// 			}
// 			defer bus.AssertExpectations(t)

// 			s := &Session{
// 				bus:       bus,
// 				store:     store,
// 				unmarshal: UnmarshalJSON,
// 			}

// 			payload, err := json.Marshal(tt.event)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			ctx := context.WithValue(context.Background(), corev2.NamespaceKey, tt.event.Entity.Namespace)
// 			if err := s.handleEvent(ctx, payload); (err != nil) != tt.wantErr {
// 				t.Errorf("Session.handleEvent() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }
