package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type testMessageType struct {
	Data string
}

func TestSendLoop(t *testing.T) {
	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive()
		msg, err := conn.Receive()

		assert.NoError(t, err)
		assert.Equal(t, "keepalive", msg.Type)
		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.NotNil(t, event.Entity)
		assert.Equal(t, "agent", event.Entity.Class)
		assert.NotEmpty(t, event.Entity.System.Hostname)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg := NewConfig()
	cfg.BackendURL = wsURL
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}
	<-done
	ta.Stop()
}

func TestReceiveLoop(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive()

		msgBytes, err := json.Marshal(testMessage)
		assert.NoError(t, err)
		tm := &transport.Message{
			Type:    "testMessageType",
			Payload: msgBytes,
		}
		err = conn.Send(tm)
		assert.NoError(t, err)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	cfg := NewConfig()
	cfg.BackendURL = wsURL
	ta := NewAgent(cfg)
	ta.addHandler("testMessageType", func(payload []byte) error {
		msg := &testMessageType{}
		err := json.Unmarshal(payload, msg)
		assert.NoError(t, err)
		assert.Equal(t, testMessage.Data, msg.Data)
		done <- struct{}{}
		return nil
	})
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}
	msgBytes, _ := json.Marshal(&testMessageType{"message"})
	ta.sendMessage("testMessageType", msgBytes)
	<-done
	<-done
	ta.Stop()
}

func TestReconnect(t *testing.T) {
	control := make(chan struct{})
	connectionCount := 0
	server := transport.NewServer()
	mutex := &sync.Mutex{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		// throw away handshake
		bhsm := &transport.Message{
			Type:    types.BackendHandshakeType,
			Payload: []byte("{}"),
		}
		conn.Send(bhsm)
		conn.Receive()
		mutex.Lock()
		connectionCount++
		mutex.Unlock()
		<-control
		conn.Close()
	}))
	defer ts.Close()

	// connect with an agent
	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	cfg := NewConfig()
	cfg.BackendURL = wsURL
	ta := NewAgent(cfg)
	err := ta.Run()
	assert.NoError(t, err)
	if err != nil {
		assert.FailNow(t, "agent failed to run")
	}
	control <- struct{}{}
	mutex.Lock()
	assert.Equal(t, 1, connectionCount)
	mutex.Unlock()

	control <- struct{}{}
	mutex.Lock()
	assert.Condition(t, func() bool { return connectionCount > 1 })
	mutex.Unlock()
	ta.Stop()
}
