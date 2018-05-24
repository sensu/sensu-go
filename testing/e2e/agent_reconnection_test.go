package e2e

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/require"
)

func TestAgentReconnection(t *testing.T) {
	t.Parallel()

	backend, cleanup, err := newBackendProcess(40005, 40006, 40007, 40008, 40009)
	if err != nil {
		t.Fatal(err)
	}

	defer cleanup()

	require.NoError(t, backend.Start())

	if !waitForBackend(backend.HTTPURL) {
		t.Fatal("backend not ready")
	}

	// Initializes sensuctl
	sensuctl, cleanup := newCustomSensuctl(t, backend.WSURL, backend.HTTPURL, "default", "default")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestAgentReconnection",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	getKeepaliveEventCmd := []string{"event", "info",
		agent.ID, "keepalive",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	}

	var output []byte
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getKeepaliveEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no keepalive received: %s", string(output))
	}

	event1 := &types.Event{}
	require.NoError(t, json.Unmarshal(output, event1))
	require.NotNil(t, event1)

	// Now terminate the backend
	require.NoError(t, backend.Terminate())

	// Restart the backend
	if err := backend.Start(); err != nil {
		log.Fatal(err)
	}

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getKeepaliveEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event2 := &types.Event{}
		if err := json.Unmarshal(output, event2); err != nil || event2 == nil {
			return false, nil
		}

		// Ensure we received a new keepalive message from the agent
		if event1.Timestamp == event2.Timestamp {
			// Let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no new keepalive received since backend was restarted: %s", string(output))
	}
}
