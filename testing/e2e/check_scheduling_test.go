package e2e

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckScheduling(t *testing.T) {
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

	// Create a check that publish check requests
	check := types.FixtureCheckConfig("TestCheckScheduling")
	check.Publish = true
	check.Interval = 1
	check.Subscriptions = []string{"test"}

	err = sensuClient.CreateCheck(check)
	assert.NoError(t, err)
	_, err = sensuClient.FetchCheck(check.Name)
	assert.NoError(t, err)

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	// Stop publishing check requests
	check.Publish = false
	err = sensuClient.UpdateCheck(check)
	assert.NoError(t, err)

	_, err = sensuClient.FetchCheck(check.Name)
	assert.NoError(t, err)

	// Give it few seconds to make sure we are not publishing check requests
	time.Sleep(20 * time.Second)

	// Retrieve the number of check results sent
	event, err := sensuClient.FetchEvent(ap.AgentID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	if event == nil {
		assert.FailNow(t, "no event was returned from the client.")
	}
	count1 := len(event.Check.History)

	// Give it few seconds to make sure we did not published additional chekc requests
	time.Sleep(10 * time.Second)

	// Retrieve (again) the number of check results sent
	event, err = sensuClient.FetchEvent(ap.AgentID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	count2 := len(event.Check.History)

	// Make sure no new check results were sent
	assert.Equal(t, count1, count2)

	// Start publishing check requests again
	check.Publish = true
	err = sensuClient.UpdateCheck(check)
	assert.NoError(t, err)

	// Give it few seconds to make sure it picks up the change
	time.Sleep(10 * time.Second)

	// Retrieve (again) the number of check results sent
	event, err = sensuClient.FetchEvent(ap.AgentID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	count3 := len(event.Check.History)

	// Make sure new check results were sent
	assert.NotEqual(t, count2, count3)
}
