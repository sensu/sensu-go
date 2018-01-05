package e2e

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
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

	// Create an authenticated HTTP Sensu client. newSensuClient is deprecated but
	// sensuctl does not currently support objects updates with flag parameters
	sensuClient := newSensuClient(backend.HTTPURL)

	// Create a check that contains a hook with status non-zero
	check := types.FixtureCheckConfig("TestCheckHooks")
	check.Command = "foo"
	check.Publish = true
	check.Interval = 1
	check.Subscriptions = []string{"test"}

	err := sensuClient.CreateCheck(check)
	assert.NoError(t, err)
	_, err = sensuClient.FetchCheck(check.Name)
	assert.NoError(t, err)

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	// Retrieve a new event
	event, err := sensuClient.FetchEvent(agent.ID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	if event == nil {
		assert.FailNow(t, "no event was returned from the client.")
	}
	// Hook hook1 does not exist, no check hook should execute
	assert.Empty(t, event.Hooks)

	// Create a hook with name hook1 which gets added to check in FixtureCheckConfig
	hook := types.FixtureHookConfig("hook1")

	err = sensuClient.CreateHook(hook)
	assert.NoError(t, err)
	_, err = sensuClient.FetchHook(hook.Name)
	assert.NoError(t, err)

	// Give it a few seconds for the check to execute with the check hook
	time.Sleep(10 * time.Second)

	// Retrieve a new event
	event, err = sensuClient.FetchEvent(agent.ID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	if event == nil {
		assert.FailNow(t, "no event was returned from the client.")
	}
	// Hook hook1 now exists, a check hook should be written to the event
	assert.NotEmpty(t, event.Hooks)
}
