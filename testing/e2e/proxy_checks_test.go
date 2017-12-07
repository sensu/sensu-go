package e2e

import (
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
	backend, cleanup := newBackend()
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestProxyChecks",
		BackendURLs: []string{backend.WSURL},
	}
	_, cleanup = newAgent(agentConfig)
	defer cleanup()

	// Create an authenticated HTTP Sensu client
	sensuClient := newSensuClient(backend.HTTPURL)

	// Create a check that specifies a source
	check := types.FixtureCheckConfig("check_router")
	check.Source = "router"
	check.Subscriptions = []string{"test"}
	check.Interval = 1

	err := sensuClient.CreateCheck(check)
	assert.NoError(t, err)
	_, err = sensuClient.FetchCheck(check.Name)
	assert.NoError(t, err)

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	// We should now have an entity that represents the source of this check
	entity, err := sensuClient.FetchEntity(check.Source)
	assert.NoError(t, err)
	assert.NotNil(t, entity)

	// We should also have an event listed under that source
	event, err := sensuClient.FetchEvent(check.Source, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
}
