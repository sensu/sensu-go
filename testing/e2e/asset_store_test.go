package e2e

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// Test asset creation -> check creation with runtime_dependency
func TestAssetStore(t *testing.T) {
	// Start the backend
	bep, cleanup := newBackendProcess()
	defer cleanup()

	err := bep.Start()
	if err != nil {
		log.Panic(err)
	}

	// Make sure the backend is available
	backendWSURL := fmt.Sprintf("ws://127.0.0.1:%d/", bep.AgentPort)
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", bep.APIPort)
	backendIsOnline := waitForBackend(backendHTTPURL)
	assert.True(t, backendIsOnline)

	// Configure the agent
	ap := &agentProcess{
		// testing the StringSlice for backend-url and the backend selector.
		BackendURLs: []string{backendWSURL, backendWSURL},
		AgentID:     "TestAssetStore",
	}

	// Start the agent
	err = ap.Start()
	if err != nil {
		log.Panic(err)
	}
	defer ap.Kill()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backendHTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Create an asset
	asset := &types.Asset{
		Name:         "asset1",
		Organization: "default",
		URL:          "http:127.0.0.1",
		Sha512:       "12345678",
	}
	output, err := sensuctl.run("asset", "create", asset.Name,
		"--organization", asset.Organization,
		"--url", asset.URL,
		"--sha512", asset.Sha512,
	)
	assert.NoError(t, err, string(output))

	// Create a check
	check := &types.CheckConfig{
		Name:          "test",
		Command:       "echo output && exit 1",
		Interval:      1,
		Subscriptions: []string{"test"},
		Handlers:      []string{"test"},
		Environment:   "default",
		Organization:  "default",
		RuntimeAssets: []string{"asset"},
	}
	output, err = sensuctl.run("check", "create", check.Name,
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--subscriptions", strings.Join(check.Subscriptions, ","),
		"--handlers", strings.Join(check.Handlers, ","),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--runtime-assets", strings.Join(check.RuntimeAssets, ","),
		"--publish",
	)
	assert.NoError(t, err, string(output))

	time.Sleep(10 * time.Second)

	// There should be a stored event
	output, err = sensuctl.run("event", "info", ap.AgentID, check.Name)
	assert.NoError(t, err, string(output))

	event := types.Event{}
	json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestAssetStore", event.Entity.ID)
	assert.Equal(t, "test", event.Check.Config.Name)
}
