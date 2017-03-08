package backend

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestGoodHandshake(t *testing.T) {
	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		session := NewSession(conn)
		err = session.Start()
		assert.NoError(t, err)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	conn, err := transport.Connect(wsURL)
	assert.NoError(t, err)
	err = conn.Send(context.TODO(), types.AgentHandshakeType, []byte("{}"))
	assert.NoError(t, err)
	msgType, m, err := conn.Receive(context.TODO())
	assert.NoError(t, err)
	assert.Equal(t, msgType, types.BackendHandshakeType)
	handshake := types.BackendHandshake{}
	err = json.Unmarshal(m, &handshake)
	assert.NoError(t, err)
	<-done
}
