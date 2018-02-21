package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TestProxyChecks ensures that the following use case is working:
// As a user, I want to run checks on dynamically created entities, so that I
// can monitor external resources
func TestProxyChecks(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Create a check that specifies a source. This check will create a new proxy
	// entity using the "proxy-entity-id" as its ID.
	source := "router"
	checkRouter := types.FixtureCheckConfig("check_router")
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
		ID:          "TestProxyChecks",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Give it few seconds to make sure we've published a check request and that
	// the proxy entity is created
	time.Sleep(10 * time.Second)

	// We should now have an entity that represents the proxy entity specified in
	// this check
	output, err = sensuctl.run("entity", "info", source)
	assert.NoError(t, err, string(output))

	// We should also have an event listed under that proxy entity with the check
	output, err = sensuctl.run("event", "info", source, checkRouter.Name)
	assert.NoError(t, err, string(output))

	// Create a proxy check request that will specifically target this proxy
	// entity
	checkProxy := types.FixtureCheckConfig("check_proxy")
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
	)
	assert.NoError(t, err, string(output))

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	// We should now have an event for our proxy check requests under that proxy
	// entity
	output, err = sensuctl.run("event", "info", source, checkProxy.Name)
	assert.NoError(t, err, string(output))

	// Make sure the agent's entity did not produce an event for our proxy check
	// request, because it shouldn't have matched the entity_attributes
	output, err = sensuctl.run("event", "info", agent.ID, checkProxy.Name)
	assert.Error(t, err, string(output))
}
