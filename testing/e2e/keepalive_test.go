package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TODO(greg): Yeah, this is really just one enormous test for all e2e stuff.
// I'd love to see this organized better.
func TestAgentKeepalives(t *testing.T) {
	// Start the backend
	bep, cleanup := newBackendProcess()
	defer cleanup()

	err := bep.Start()
	if err != nil {
		log.Panic(err)
	}

	backendWSURL := fmt.Sprintf("ws://127.0.0.1:%d/", bep.AgentPort)
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", bep.APIPort)

	// Make sure the backend is available
	backendIsOnline := waitForBackend(backendHTTPURL)
	assert.True(t, backendIsOnline)

	// Configure the agent
	ap := &agentProcess{
		// testing the StringSlice for backend-url and the backend selector.
		BackendURLs: []string{backendWSURL, backendWSURL},
		AgentID:     "TestKeepalives",
	}

	err = ap.Start()
	assert.NoError(t, err)

	defer func() {
		bep.Kill()
		ap.Kill()
	}()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backendHTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Retrieve the entitites
	output, err := sensuctl.run("entity", "list")
	assert.NoError(t, err)

	entities := []types.Entity{}
	json.Unmarshal(output, &entities)

	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "TestKeepalives", entities[0].ID)
	assert.Equal(t, "agent", entities[0].Class)
	assert.NotEmpty(t, entities[0].System.Hostname)
	assert.NotZero(t, entities[0].LastSeen)

	falsePath := testutil.CommandPath(filepath.Join(binDir, "false"))
	falseAbsPath, err := filepath.Abs(falsePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, falseAbsPath)

	// Create a standard check
	checkName := "test_check"
	_, err = sensuctl.run("check", "create", checkName,
		"--command", falseAbsPath,
		"--interval", "1",
		"--subscriptions", "test",
	)
	assert.NoError(t, err)

	// Make sure the check has been properly created
	output, err = sensuctl.run("check", "info", checkName)
	assert.NoError(t, err)

	result := types.CheckConfig{}
	json.Unmarshal(output, &result)
	assert.Equal(t, result.Name, checkName)

	time.Sleep(30 * time.Second)

	// At this point, we should have 21 failing status codes for testcheck2
	output, err = sensuctl.run("event", "info", ap.AgentID, checkName)
	assert.NoError(t, err)

	event := types.Event{}
	json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestKeepalives", event.Entity.ID)
	assert.Equal(t, checkName, event.Check.Config.Name)
	// TODO(greg): ensure results are as expected.

	// Test the agent HTTP API
	newEvent := types.FixtureEvent(ap.AgentID, "proxy-check")
	encoded, _ := json.Marshal(newEvent)
	url := fmt.Sprintf("http://127.0.0.1:%d/events", ap.APIPort)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(encoded))

	client := &http.Client{}
	res, err := client.Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()

	// Give it a second to receive the new event
	time.Sleep(5 * time.Second)

	// Make sure the new event has been received
	output, err = sensuctl.run("event", "info", ap.AgentID, "proxy-check")
	assert.NoError(t, err, string(output))
	assert.NotNil(t, output)
}
