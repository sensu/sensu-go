package agentd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGoodHandshake(t *testing.T) {
	done := make(chan struct{})
	server := transport.NewServer()
	var session *Session

	st := &mockstore.MockStore{}
	st.On("UpdateEntity", mock.AnythingOfType("*types.Entity")).Return(nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		bus := &messaging.WizardBus{}
		bus.Start()
		session, err = NewSession(conn, bus, st)
		assert.NoError(t, err)
		err = session.Start()
		assert.NoError(t, err)
		done <- struct{}{}
		session.Stop()
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	conn, err := transport.Connect(wsURL)
	assert.NoError(t, err)

	entity := &types.AgentHandshake{
		ID:            "id",
		Subscriptions: []string{"subscription1"},
	}
	payload, err := json.Marshal(entity)
	assert.NoError(t, err)
	msg := &transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: payload,
	}
	err = conn.Send(msg)
	assert.NoError(t, err)
	resp, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, types.BackendHandshakeType, resp.Type)
	handshake := types.BackendHandshake{}
	err = json.Unmarshal(resp.Payload, &handshake)
	assert.NoError(t, err)
	<-done
}
