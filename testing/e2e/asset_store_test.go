package e2e

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test asset creation -> check creation with runtime_dependency
func TestAssetStore(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestAssetStore",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create an asset
	asset := &types.Asset{
		Name:         "asset1",
		Organization: "default",
		URL:          "http:foo.com",
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
	output, err = sensuctl.run("event", "info", agent.ID, check.Name)
	assert.NoError(t, err, string(output))

	event := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestAssetStore", event.Entity.ID)
	assert.Equal(t, "test", event.Check.Name)
	assert.Equal(t, "asset", strings.Join(event.Check.RuntimeAssets, ","))
}
