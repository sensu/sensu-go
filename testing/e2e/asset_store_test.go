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

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestAssetStore",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create an asset
	asset := &types.Asset{
		Name:         "asset1",
		Organization: agent.Organization,
		URL:          "http:foo.com",
		Sha512:       "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17",
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
		Environment:   agent.Environment,
		Organization:  agent.Organization,
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
	output, err = sensuctl.run("event", "info", agent.ID, check.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
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
