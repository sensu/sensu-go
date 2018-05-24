package e2e

import (
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TestProxyChecks ensures that the following use case is working:
// As a user, I want to run checks on dynamically created entities, so that I
// can monitor external resources
func TestProxyChecks(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Create a check that specifies a source. This check will create a new proxy
	// entity using the "proxy-entity-id" as its ID.
	source := "router"
	checkRouter := types.FixtureCheckConfig("check_router")
	checkRouter.Organization = sensuctl.Organization
	checkRouter.Environment = sensuctl.Environment
	output, err := sensuctl.run("check", "create", checkRouter.Name,
		"--command", checkRouter.Command,
		"--interval", "1",
		"--subscriptions", "test",
		"--handlers", strings.Join(checkRouter.Handlers, ","),
		"--organization", checkRouter.Organization,
		"--environment", checkRouter.Environment,
		"--runtime-assets", strings.Join(checkRouter.RuntimeAssets, ","),
		"--proxy-entity-id", "router",
		"--publish",
	)
	assert.NoError(t, err, string(output))

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestProxyChecks",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", source, checkRouter.Name,
			"--organization", sensuctl.Organization,
			"--environment", sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no check event received: %s", string(output))
	}

	// Create a proxy check request that will specifically target this proxy
	// entity
	checkProxy := types.FixtureCheckConfig("check_proxy")
	checkProxy.Organization = sensuctl.Organization
	checkProxy.Environment = sensuctl.Environment
	output, err = sensuctl.run("check", "create", checkProxy.Name,
		"--command", checkProxy.Command,
		"--interval", "1",
		"--subscriptions", "test",
		"--handlers", strings.Join(checkProxy.Handlers, ","),
		"--organization", checkProxy.Organization,
		"--environment", checkProxy.Environment,
		"--runtime-assets", strings.Join(checkProxy.RuntimeAssets, ","),
		"--publish",
	)
	assert.NoError(t, err, string(output))

	// Set the proxy check requests to our previous check
	proxyRequests := []byte(`{"entity_attributes": ["entity.Class == \"proxy\""]}`)
	path, cleanup := writeTempFile(t, proxyRequests, "proxyrequests.json")
	defer cleanup()
	output, err = sensuctl.run("check", "set-proxy-requests", checkProxy.Name,
		"-f", path,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	assert.NoError(t, err, string(output))

	// Give it few seconds to make sure we've published a check request
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", source, checkProxy.Name,
			"--organization", sensuctl.Organization,
			"--environment", sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event for the proxy request received: %s", string(output))
	}

	// Make sure the agent's entity did not produce an event for our proxy check
	// request, because it shouldn't have matched the entity_attributes
	output, err = sensuctl.run(
		"event", "info", agent.ID, checkProxy.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	assert.Error(t, err, string(output))
}
