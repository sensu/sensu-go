package agentd

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
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

func (t *testTransport) Reconnect(wsServerURL string, tlsOpts *types.TLSOptions, requestHeader http.Header) error {
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
	).Return(&types.Namespace{}, nil)

	cfg := SessionConfig{
		AgentName:     "testing",
		Namespace:     "acme",
		Subscriptions: []string{"testing"},
	}
	session, err := NewSession(cfg, conn, bus, st)
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
	).Return(&types.Namespace{}, fmt.Errorf("error"))

	cfg := SessionConfig{
		Subscriptions: []string{"testing"},
	}
	session, err := NewSession(cfg, conn, bus, st)
	assert.Nil(t, session)
	assert.Error(t, err)
}
