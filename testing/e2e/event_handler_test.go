package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test check event creation -> event handler.
func TestEventHandler(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestEventHandler",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	handlerJSONFile := fmt.Sprintf("%s/TestEventHandler%v", os.TempDir(), os.Getpid())

	// Create a handler
	handler := &types.Handler{
		Name:         "test",
		Type:         "pipe",
		Command:      fmt.Sprintf("cat > %s", handlerJSONFile),
		Environment:  agent.Environment,
		Organization: agent.Organization,
	}
	output, err := sensuctl.run("handler", "create", handler.Name,
		"--type", handler.Type,
		"--command", handler.Command,
		"--organization", handler.Organization,
		"--environment", handler.Environment,
	)
	assert.NoError(t, err, string(output))

	// Create a check
	check := &types.CheckConfig{
		Name:          "test",
		Command:       "echo output && exit 1",
		Interval:      1,
		Subscriptions: []string{"test"},
		Handlers:      []string{"test"},
		Environment:   agent.Environment,
		Organization:  agent.Organization,
	}
	output, err = sensuctl.run("check", "create", check.Name,
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--subscriptions", strings.Join(check.Subscriptions, ","),
		"--handlers", strings.Join(check.Handlers, ","),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", agent.ID, check.Name,
			"--organization", sensuctl.Organization,
			"--environment", sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event received: %s", string(output))
	}

	event := &types.Event{}
	require.NoError(t, json.Unmarshal(output, event))
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestEventHandler", event.Entity.ID)
	assert.Equal(t, "test", event.Check.Name)

	// There should be a JSON event file in the OS temp directory
	_, err = os.Stat(handlerJSONFile)
	assert.NoError(t, err)
}
