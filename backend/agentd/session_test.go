package agentd

import (
	"fmt"
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

	bus := &messaging.WizardBus{}
	require.NoError(t, bus.Start())

	st := &mockstore.MockStore{}
	st.On(
		"GetEnvironment",
		mock.Anything,
		"org",
		"env",
	).Return(&types.Environment{}, nil)

	cfg := SessionConfig{
		AgentID:       "testing",
		Organization:  "org",
		Environment:   "env",
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

	bus := &messaging.WizardBus{}
	require.NoError(t, bus.Start())

	st := &mockstore.MockStore{}
	st.On(
		"UpdateEntity",
		mock.Anything,
		mock.AnythingOfType("*types.Entity"),
	).Return(nil)
	st.On(
		"GetEnvironment",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(&types.Environment{}, fmt.Errorf("error"))

	cfg := SessionConfig{
		Subscriptions: []string{"testing"},
	}
	session, err := NewSession(cfg, conn, bus, st)
	assert.Nil(t, session)
	assert.Error(t, err)
}
