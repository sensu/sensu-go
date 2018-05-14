package e2e

import (
	"encoding/json"
	"log"
	"testing"
	"time"

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

	// Give it few seconds to make sure the agent sent a keepalive
	time.Sleep(10 * time.Second)

	// Retrieve the event for keepalive
	output, err := sensuctl.run(
		"event", "info", agent.ID, "keepalive",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err, string(output))

	event1 := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event1))
	require.NotNil(t, event1)

	// Now terminate the backend
	require.NoError(t, backend.Terminate())

	// Restart the backend
	if err := backend.Start(); err != nil {
		log.Fatal(err)
	}

	// Give it few seconds to make sure the agent sent a keepalive
	time.Sleep(10 * time.Second)

	// Retrieve the the latest event for keepalive
	output, err = sensuctl.run(
		"event", "info", agent.ID, "keepalive",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err, string(output))

	event2 := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event2))
	require.NotNil(t, event2)

	// Ensure we received a new keepalive message from the agent
	require.NotEqual(t, event1.Timestamp, event2.Timestamp)
}
