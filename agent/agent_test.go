package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
)

type testMessageType struct {
	Data string
}

func TestSendLoop(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transport, err := server.Serve(w, r)
		assert.NoError(t, err)
		msgType, payload, err := transport.Receive(context.TODO())

		assert.NoError(t, err)
		assert.Equal(t, "testMessageType", msgType)
		m := &testMessageType{"message"}
		assert.NoError(t, json.Unmarshal(payload, m))
		assert.Equal(t, testMessage.Data, m.Data)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	ta := NewAgent(Config{
		BackendURL: wsURL,
	})
	err := ta.Run()
	assert.NoError(t, err)
	msgBytes, _ := json.Marshal(&testMessageType{"message"})
	ta.sendMessage("testMessageType", msgBytes)
	<-done
}

func TestReceiveLoop(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		transport, err := server.Serve(w, r)
		assert.NoError(t, err)
		msgBytes, err := json.Marshal(testMessage)
		assert.NoError(t, err)
		err = transport.Send(context.Background(), "testMessageType", msgBytes)
		assert.NoError(t, err)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	ta := NewAgent(Config{
		BackendURL: wsURL,
	})
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
	msgBytes, _ := json.Marshal(&testMessageType{"message"})
	ta.sendMessage("testMessageType", msgBytes)
	<-done
	<-done
}

func TestReconnect(t *testing.T) {
	control := make(chan struct{}, 1)
	connectionCount := 0
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		assert.NoError(t, err)
		connectionCount++
		control <- struct{}{}
		<-control
		conn.Close()
		control <- struct{}{}
	}))
	defer ts.Close()

	// connect with an agent
	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	ta := NewAgent(Config{
		BackendURL: wsURL,
	})
	err := ta.Run()
	assert.NoError(t, err)
	<-control
	assert.Equal(t, 1, connectionCount)
	control <- struct{}{}
	// that'll disconnect it
	time.Sleep(1 * time.Second)
	assert.True(t, ta.disconnected)
	time.Sleep(2 * time.Second)
	assert.False(t, ta.disconnected)

}
