package e2e

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TestProxyChecks ensures that the following user case is working:
// As a user, I want to run checks on dynamically created entities, so that I
// can monitor external resources
func TestProxyChecks(t *testing.T) {
	// Start the backend
	bep, cleanup := newBackendProcess()
	defer cleanup()

	err := bep.Start()
	if err != nil {
		log.Panic(err)
	}
	defer bep.Kill()

	// Make sure the backend is available
	backendWSURL := fmt.Sprintf("ws://127.0.0.1:%d/", bep.AgentPort)
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", bep.APIPort)
	backendIsOnline := waitForBackend(backendHTTPURL)
	assert.True(t, backendIsOnline)

	// Configure the agent
	ap := &agentProcess{
		// testing the StringSlice for backend-url and the backend selector.
		BackendURLs: []string{backendWSURL, backendWSURL},
		AgentID:     "TestCheckScheduling",
	}

	// Start the agent
	err = ap.Start()
	if err != nil {
		log.Panic(err)
	}
	defer ap.Kill()

	// Give it few seconds to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	// Create an authenticated HTTP Sensu client
	sensuClient := newSensuClient(backendHTTPURL)

	// Create a check that specifies a source
	check := types.FixtureCheckConfig("check_router")
	check.Source = "router"
	check.Subscriptions = []string{"test"}
	check.Interval = 1

	err = sensuClient.CreateCheck(check)
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
