// +build integration

package agent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMessageType struct {
	Data string
}

func TestSendLoop(t *testing.T) {
	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		require.NoError(t, err)

		msg, err := conn.Receive()
		assert.NoError(t, err)
		assert.Equal(t, "keepalive", msg.Type)

		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.NotNil(t, event.Entity)
		assert.Equal(t, "agent", event.Entity.EntityClass)
		assert.NotEmpty(t, event.Entity.System)
		done <- struct{}{}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)
	err := ta.Run()
	require.NoError(t, err)
	defer ta.Stop()
	<-done
}

func TestReceiveLoop(t *testing.T) {
	testMessage := &testMessageType{"message"}

	done := make(chan struct{})
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		require.NoError(t, err)

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
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)
	ta.handler.AddHandler("testMessageType", func(payload []byte) error {
		msg := &testMessageType{}
		err := json.Unmarshal(payload, msg)
		assert.NoError(t, err)
		assert.Equal(t, testMessage.Data, msg.Data)
		done <- struct{}{}
		return nil
	})
	err := ta.Run()
	require.NoError(t, err)
	defer ta.Stop()
	msgBytes, _ := json.Marshal(&testMessageType{"message"})
	ta.sendMessage("testMessageType", msgBytes)
	<-done
	<-done
}

// TestPeriodicKeepalive checks that a running Agent sends its periodic
// keepalive messages at the expected frequency, allowing for +/- 2s drift.
func TestPeriodicKeepalive(t *testing.T) {
	done := make(chan struct{})

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	server := transport.NewServer()

	testKeepalive := func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		require.NoError(t, err)

		lastKeepalive := time.Time{}
		keepaliveInterval := time.Duration(cfg.KeepaliveInterval) * time.Second

		for keepaliveCount := 0; keepaliveCount < 10; keepaliveCount++ {
			msg, err := conn.Receive()
			assert.NoError(t, err)

			if msg.Type == "keepalive" {
				if keepaliveCount > 0 {
					expected := lastKeepalive.Add(keepaliveInterval)
					actual := mockTime.Now()
					assert.WithinDuration(t, expected, actual, (2 * time.Second))
				}
				lastKeepalive = mockTime.Now()
			}
		}

		conn.Close()
		done <- struct{}{}
	}

	ts := httptest.NewServer(http.HandlerFunc(testKeepalive))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	cfg.BackendURLs = []string{wsURL}

	mockTime.Start()
	defer mockTime.Stop()

	ta := NewAgent(cfg)
	err := ta.Run()
	require.NoError(t, err)
	defer ta.Stop()

	<-done
}

func TestKeepaliveLoggingRedaction(t *testing.T) {
	errors := make(chan error, 100)
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			close(errors)
		}()
		conn, err := server.Serve(w, r)
		require.NoError(t, err)

		msg, err := conn.Receive()
		assert.NoError(t, err)
		assert.Equal(t, "keepalive", msg.Type)

		event := &types.Event{}
		assert.NoError(t, json.Unmarshal(msg.Payload, event))
		assert.NotNil(t, event.Entity)
		assert.Equal(t, "agent", event.Entity.EntityClass)
		assert.NotEmpty(t, event.Entity.System)

		// Make sure the ec2_access_key attribute is redacted, which indicates it was
		// received as such in keepalives
		label := event.Entity.Labels["ec2_access_key"]
		if got, want := label, types.Redacted; got != want {
			errors <- fmt.Errorf("%q != %q", got, want)
		}

		label = event.Entity.Labels["secret"]
		if got, want := label, types.Redacted; got == want {
			errors <- fmt.Errorf("secret was redacted")
		}
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.AgentName = "TestLoggingRedaction"
	cfg.Labels = map[string]string{"ec2_access_key": "P@ssw0rd!", "secret": "P@ssw0rd!"}
	cfg.Redact = []string{"ec2_access_key"}
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)
	err := ta.Run()
	require.NoError(t, err)
	defer ta.Stop()
	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
}

func TestInvalidAgentName_GH2022(t *testing.T) {
	server := transport.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Serve(w, r)
		require.NoError(t, err)
		conn.Close()
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.AgentName = "Test Agent"
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	ta := NewAgent(cfg)
	err := ta.Run()
	require.Error(t, err)
	defer ta.Stop()
}
