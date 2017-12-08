package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TestProxyChecks ensures that the following user case is working:
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

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestProxyChecks",
		BackendURLs: []string{backend.WSURL},
	}
	_, cleanup = newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that specifies a source
	source := "router"
	check := types.FixtureCheckConfig("check_router")
	output, err := sensuctl.run("check", "create", check.Name,
		"--command", check.Command,
		"--interval", "1",
		"--subscriptions", "test",
		"--handlers", strings.Join(check.Handlers, ","),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--runtime-assets", strings.Join(check.RuntimeAssets, ","),
		"--source", "router",
		"--publish",
	)
	assert.NoError(t, err, string(output))

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	// We should now have an entity that represents the source of this check
	output, err = sensuctl.run("entity", "info", source)
	assert.NoError(t, err, string(output))

	// We should also have an event listed under that source
	// There should be a stored event
	output, err = sensuctl.run("event", "info", source, check.Name)
	assert.NoError(t, err, string(output))
}
