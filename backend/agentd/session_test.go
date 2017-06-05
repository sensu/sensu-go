package agentd

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testTransport struct {
	sendCh chan *transport.Message
	closed bool
}

func (t testTransport) Closed() bool {
	return t.closed
}

func (t testTransport) Close() error {
	t.closed = true
	return nil
}

func (t testTransport) Send(msg *transport.Message) error {
	t.sendCh <- msg
	return nil
}

func (t testTransport) Receive() (*transport.Message, error) {
	return <-t.sendCh, nil
}

func TestGoodHandshake(t *testing.T) {
	conn := testTransport{
		sendCh: make(chan *transport.Message, 10),
	}

	bus := &messaging.WizardBus{}
	bus.Start()

	st := &mockstore.MockStore{}
	st.On("UpdateEntity", mock.AnythingOfType("*types.Entity")).Return(nil)

	session, err := NewSession(conn, bus, st)
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "unable to create session")
	}
	assert.NotNil(t, session)

	hsBytes, _ := json.Marshal(&types.AgentHandshake{})
	conn.Send(&transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: hsBytes,
	})
	assert.NoError(t, session.Start())
}
