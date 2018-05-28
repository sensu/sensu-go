package e2e

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckHooks(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestCheckHooks",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that contains a hook with status non-zero
	check := types.FixtureCheckConfig("TestCheckHooks")
	check.Organization = sensuctl.Organization
	check.Environment = sensuctl.Environment
	check.Command = "foo"
	check.Publish = true
	check.Interval = 5
	check.Subscriptions = []string{"test"}

	output, err := sensuctl.run("check", "create", check.Name,
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--runtime-assets", strings.Join(check.RuntimeAssets, ","),
		"--subscriptions", strings.Join(check.Subscriptions, ","),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	output, err = sensuctl.run("check", "info", check.Name)
	assert.NoError(t, err, string(output))

	getCheckEventCmd := []string{"event", "info", agent.ID, check.Name}

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event received: %s", string(output))
	}

	// Retrieve a new event
	event := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NoError(t, err)
	assert.NotNil(t, event)

	// Hook hook1 does not exist, no check hook should execute
	assert.Empty(t, event.Check.Hooks)

	// Create a hook with hook name hook1
	hook := types.FixtureHookConfig("hook1")
	hook.Organization = sensuctl.Organization
	hook.Environment = sensuctl.Environment
	hook.Command = "echo {{ .ID }}"

	output, err = sensuctl.run("hook", "create", hook.Name,
		"--command", hook.Command,
		"--organization", hook.Organization,
		"--environment", hook.Environment,
	)
	assert.NoError(t, err, string(output))

	output, err = sensuctl.run("hook", "info", hook.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err, string(output))

	// Add hook with hook name hook1 to check
	checkHook := types.FixtureHookList("hook1")
	output, err = sensuctl.run("check", "set-hooks", check.Name,
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--type", checkHook.Type,
		"--hooks", strings.Join(checkHook.Hooks, ","),
	)
	assert.NoError(t, err, string(output))

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		if len(event.Check.Hooks) == 0 {
			return false, nil
		}

		// Ensure the event's assets are present
		if !strings.Contains(event.Check.Hooks[0].Output, agent.ID) {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no new event with hooks received: %s", string(output))
	}
}
