package e2e

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckHooks(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestCheckHooks",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that contains a hook with status non-zero
	check := types.FixtureCheckConfig("TestCheckHooks")
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

	// Give it few seconds to make sure we've published a check request
	time.Sleep(20 * time.Second)

	// There should be a stored event
	output, err = sensuctl.run("event", "info", agent.ID, check.Name)
	assert.NoError(t, err, string(output))

	// Retrieve a new event
	event := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NoError(t, err)
	assert.NotNil(t, event)

	// Hook hook1 does not exist, no check hook should execute
	assert.Empty(t, event.Hooks)

	// Create a hook with hook name hook1
	hook := types.FixtureHookConfig("hook1")
	hook.Command = "echo {{ .ID }}"

	output, err = sensuctl.run("hook", "create", hook.Name,
		"--command", hook.Command,
	)
	assert.NoError(t, err, string(output))

	output, err = sensuctl.run("hook", "info", hook.Name)
	assert.NoError(t, err, string(output))

	// Add hook with hook name hook1 to check
	checkHook := types.FixtureHookList("hook1")
	output, err = sensuctl.run("check", "set-hooks", check.Name,
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--type", checkHook.Type,
		"--hooks", strings.Join(checkHook.Hooks, ","),
	)
	assert.NoError(t, err, string(output))

	// Give it a few seconds for the check to execute with the check hook
	time.Sleep(20 * time.Second)

	// There should be a stored event
	output, err = sensuctl.run("event", "info", agent.ID, check.Name)
	assert.NoError(t, err, string(output))

	// Retrieve a new event
	event = types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NoError(t, err)
	assert.NotNil(t, event)

	// Ensure the token substitution has been applied for the hook's command
	assert.Contains(t, event.Hooks[0].Output, agent.ID)

	// Hook hook1 now exists, a check hook should be written to the event
	assert.NotEmpty(t, event.Hooks)
}
