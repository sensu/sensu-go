package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type testStore struct{}

func (t *testStore) GetEntityByID(string) (*types.Entity, error) {
	return &types.Entity{}, nil
}
func (t *testStore) UpdateEntity(*types.Entity) error {
	return nil
}
func (t *testStore) DeleteEntity(*types.Entity) error {
	return nil
}

func TestGoodHandshake(t *testing.T) {
	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		session := NewSession(conn, &testStore{})
		err = session.Start()
		assert.NoError(t, err)
		done <- struct{}{}
		session.Stop()
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	conn, err := transport.Connect(wsURL)
	assert.NoError(t, err)
	msg := &transport.Message{
		Type:    types.AgentHandshakeType,
		Payload: []byte("{}"),
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
