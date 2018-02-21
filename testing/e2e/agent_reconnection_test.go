package e2e

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentReconnection(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestAgentReconnection",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Give it few seconds to make sure the agent sent a keepalive
	time.Sleep(10 * time.Second)

	// Retrieve the event for keepalive
	output, err := sensuctl.run("event", "info", agent.ID, "keepalive")
	assert.NoError(t, err, string(output))

	event1 := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event1))
	assert.NotNil(t, event1)

	// Now kill the backend
	require.NoError(t, backend.Kill())

	// Restart the backend
	if err := backend.Start(); err != nil {
		log.Panic(err)
	}

	// Give it few seconds to make sure the agent sent a keepalive
	time.Sleep(10 * time.Second)

	// Retrieve the the latest event for keepalive
	output, err = sensuctl.run("event", "info", agent.ID, "keepalive")
	assert.NoError(t, err, string(output))

	event2 := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event2))
	assert.NotNil(t, event2)

	// Ensure we received a new keepalive message from the agent
	assert.NotEqual(t, event1.Timestamp, event2.Timestamp)
}
