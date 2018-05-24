package e2e

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/require"
)

func TestMetricExtraction(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestMetricExtraction",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that publish check requests
	output, err := sensuctl.run(
		"check", "create", "nagios-metric",
		"--publish",
		"--interval", "1",
		"--subscriptions", "test",
		"--command", `echo "PING ok - Packet loss = 0% | percent_packet_loss=0"`,
		"--output-metric-format", "nagios_perfdata",
		"--environment", sensuctl.Environment,
		"--organization", sensuctl.Organization,
	)
	require.NoError(t, err, string(output))

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", agent.ID, "nagios-metric",
			"--environment", sensuctl.Environment,
			"--organization", sensuctl.Organization); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		if len(event.Metrics.Points) == 0 || event.Metrics.Points[0].Name != "percent_packet_loss" || event.Metrics.Points[0].Value != 0.0 {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no metric event received: %s", string(output))
	}
}
