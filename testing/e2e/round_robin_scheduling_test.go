package e2e

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinScheduling(t *testing.T) {
	t.Parallel()

	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Two agents belong to backend A, one belongs to backend B
	agentCfgA := agentConfig{
		ID: "agentA",
	}
	agentCfgB := agentConfig{
		ID: "agentB",
	}
	agentCfgC := agentConfig{
		ID: "agentC",
	}

	agentA, cleanup := newAgent(agentCfgA, sensuctl, t)
	defer cleanup()

	agentB, cleanup := newAgent(agentCfgB, sensuctl, t)
	defer cleanup()

	agentC, cleanup := newAgent(agentCfgC, sensuctl, t)
	defer cleanup()

	// Create a check that publish check requests
	_, err := sensuctl.run("check", "create", t.Name(),
		"--command", testutil.CommandPath(filepath.Join(toolsDir, "true")),
		"--interval", "1",
		"--subscriptions", "test",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
		"--round-robin",
		"--publish",
	)
	require.NoError(t, err)
	_, err = sensuctl.run(
		"check", "info", t.Name(),
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)

	// Allow checks to be published
	time.Sleep(20 * time.Second)

	out, err := sensuctl.run(
		"event", "info", agentA.ID, t.Name(),
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	var eventA types.Event
	require.NoError(t, json.Unmarshal(out, &eventA))

	out, err = sensuctl.run(
		"event", "info", agentB.ID, t.Name(),
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	var eventB types.Event
	require.NoError(t, json.Unmarshal(out, &eventB))

	out, err = sensuctl.run("event", "info", agentC.ID, t.Name(),
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	var eventC types.Event
	require.NoError(t, json.Unmarshal(out, &eventC))

	histories := append(eventA.Check.History, eventB.Check.History...)
	histories = append(histories, eventC.Check.History...)

	executed := make(map[int64]struct{})
	for _, h := range histories {
		assert.Equal(t, uint32(0), h.Status)
		e := h.Executed
		executed[e] = struct{}{}
	}
	// Ensure that all executed checks have been executed at a separate time.
	// TODO(echlebek): do this with unique identifiers per check request msg.
	assert.Equal(t, len(histories), len(executed))
}
