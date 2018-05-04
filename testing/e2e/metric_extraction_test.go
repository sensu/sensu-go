package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricExtraction(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestMetricExtraction",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that publish check requests
	out, err := sensuctl.run(
		"check", "create", "nagios-metric",
		"--publish",
		"--interval", "1",
		"--subscriptions", "test",
		"--command", `echo "PING ok - Packet loss = 0% | percent_packet_loss=0"`,
		"--metric-format", "nagios_perfdata",
	)
	require.NoError(t, err, string(out))

	// FIXME: Give it few seconds to make sure we're not publishing check requests.
	time.Sleep(15 * time.Second)

	// There should be a stored event for our metric
	out, err = sensuctl.run("event", "info", agent.ID, "nagios-metric")
	assert.NoError(t, err, string(out))

	event := types.Event{}
	require.NoError(t, json.Unmarshal(out, &event))
	assert.NotNil(t, event)
	assert.NotZero(t, len(event.Metrics.Points))
	assert.Equal(t, "percent_packet_loss", event.Metrics.Points[0].Name)
	assert.Equal(t, 0.0, event.Metrics.Points[0].Value)
}
