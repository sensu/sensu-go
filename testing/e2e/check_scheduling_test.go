package e2e

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestCheckScheduling(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:                "TestCheckScheduling",
		BackendURLs:       []string{backend.WSURL},
		KeepaliveInterval: 1,
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create an authenticated HTTP Sensu client. newSensuClient is deprecated but
	// sensuctl does not currently support objects updates with flag parameters
	sensuClient := newSensuClient(backend.HTTPURL)

	// Create a check that publish check requests
	check := types.FixtureCheckConfig("TestCheckScheduling")
	check.Publish = true
	check.Interval = 1
	check.Subscriptions = []string{"test"}

	err := sensuClient.CreateCheck(check)
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
	event, err := sensuClient.FetchEvent(agent.ID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	if event == nil {
		assert.FailNow(t, "no event was returned from the client.")
	}
	count1 := len(event.Check.History)

	// Give it few seconds to make sure we did not published additional check requests
	time.Sleep(10 * time.Second)

	// Retrieve (again) the number of check results sent
	event, err = sensuClient.FetchEvent(agent.ID, check.Name)
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
	event, err = sensuClient.FetchEvent(agent.ID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	count3 := len(event.Check.History)

	// Make sure new check results were sent
	assert.NotEqual(t, count2, count3)

	// Change the check schedule to cron
	check.Interval = 0
	check.Cron = "* * * * *"
	err = sensuClient.UpdateCheck(check)
	assert.NoError(t, err)

	// Give it few seconds to make sure it picks up the change
	time.Sleep(60 * time.Second)

	// Retrieve (again) the number of check results sent
	event, err = sensuClient.FetchEvent(agent.ID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	count4 := len(event.Check.History)

	// Make sure new check results were sent
	assert.NotEqual(t, count3, count4)
}
