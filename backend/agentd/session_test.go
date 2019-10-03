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
	session, err := NewSession(cfg, conn, bus, st, UnmarshalJSON, MarshalJSON)
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
	session, err := NewSession(cfg, conn, bus, st, proto.Unmarshal, proto.Marshal)
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
	session, err := NewSession(cfg, conn, bus, st, UnmarshalJSON, MarshalJSON)
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
	conn.On("Send", mock.Anything).Return(transport.ConnectionError{"some horrible network outage"})

	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if err := bus.Start(); err != nil {
		t.Fatal(err)
	}

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
	session, err := NewSession(cfg, conn, bus, st, UnmarshalJSON, MarshalJSON)
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
	case <-session.stopping:
	case <-time.After(time.Second * 5):
		t.Fatal("broken session never stopped")
	}
}
