package agentd

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testTransport struct {
	sendCh  chan *transport.Message
	closed  bool
	sendErr error
	recvErr error
}

func (t testTransport) Closed() bool {
	return t.closed
}

func (t testTransport) Close() error {
	t.closed = true
	return nil
}

func (t testTransport) Send(msg *transport.Message) error {
	if t.sendErr != nil {
		return t.sendErr
	}
	t.sendCh <- msg
	return nil
}

func (t testTransport) Receive() (*transport.Message, error) {
	if t.recvErr != nil {
		return nil, t.recvErr
	}
	return <-t.sendCh, nil
}

func TestGoodHandshake(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

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
	).Return(&types.Environment{}, nil)

	session, err := NewSession(conn, bus, st)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to create session")
	}
	assert.NotNil(t, session)

	hsBytes, _ := json.Marshal(&types.AgentHandshake{
		Subscriptions: []string{"testing"},
	})
	conn.Send(&transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: hsBytes,
	})
	assert.NoError(t, session.Start())
}

func TestBadHandshake(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

	st := &mockstore.MockStore{}

	session, err := NewSession(conn, bus, st)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to create session")
	}
	assert.NotNil(t, session)

	conn.Send(&transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: []byte("..."),
	})
	assert.Error(t, session.Start())
}

func TestBadOrganizationHandshake(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

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

	session, _ := NewSession(conn, bus, st)
	hsBytes, _ := json.Marshal(&types.AgentHandshake{
		Subscriptions: []string{"testing"},
	})
	conn.Send(&transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: hsBytes,
	})

	assert.Error(t, session.Start())
}

func TestNoHandshake(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

	st := &mockstore.MockStore{}

	session, err := NewSession(conn, bus, st)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to create session")
	}
	assert.NotNil(t, session)

	conn.Send(&transport.Message{
		Type:    transport.MessageTypeEvent,
		Payload: []byte("..."),
	})
	assert.Error(t, session.Start())
}

func TestSendError(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

	st := &mockstore.MockStore{}

	session, err := NewSession(conn, bus, st)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to create session")
	}
	assert.NotNil(t, session)

	conn.sendErr = errors.New("error")
	assert.Error(t, session.Start())
}

func TestReceiveError(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

	st := &mockstore.MockStore{}

	session, err := NewSession(conn, bus, st)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to create session")
	}
	assert.NotNil(t, session)

	conn.recvErr = errors.New("error")
	assert.Error(t, session.Start())
}
