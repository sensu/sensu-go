// +build integration

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMessageType struct {
	Data string
}

func TestSendLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
		cancel()
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)

	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	ta, err := NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	mockTime.Start()
	defer mockTime.Stop()
	err = ta.Run(ctx)
	require.NoError(t, err)
}

func TestReceiveLoop(t *testing.T) {
	testMessage := &testMessageType{"message"}

	server := transport.NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
		cancel()
	}))
	defer ts.Close()

	wsURL := strings.Replace(ts.URL, "http", "ws", 1)
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	cfg.BackendURLs = []string{wsURL}
	cfg.API.Port = 0
	cfg.Socket.Port = 0
	ta, err := NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ta.handler.AddHandler("testMessageType", func(ctx context.Context, payload []byte) error {
		msg := &testMessageType{}
		err := json.Unmarshal(payload, msg)
		assert.NoError(t, err)
		assert.Equal(t, testMessage.Data, msg.Data)
		cancel()
		return nil
	})
	msgBytes, _ := json.Marshal(&testMessageType{"message"})
	tm := &transport.Message{Payload: msgBytes, Type: "testMessageType"}
	ta.sendMessage(tm)
	err = ta.Run(ctx)
	require.NoError(t, err)
}

func TestKeepaliveLoggingRedaction(t *testing.T) {
	errors := make(chan error, 100)
	server := transport.NewServer()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := ctx.Err(); err != nil {
			return
		}
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

		cancel()
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
	ta, err := NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	mockTime.Start()
	defer mockTime.Stop()
	err = ta.Run(ctx)
	close(errors)
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
	ta, err := NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	err = ta.Run(context.Background())
	require.Error(t, err)
}
