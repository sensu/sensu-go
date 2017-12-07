package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// Test check event creation -> event handler.
func TestEventHandler(t *testing.T) {
	// Start the backend
	backend, cleanup := newBackend()
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestEventHandler",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	handlerJSONFile := fmt.Sprintf("%s/TestEventHandler%v", os.TempDir(), os.Getpid())

	// Create a handler
	handler := &types.Handler{
		Name:         "test",
		Type:         "pipe",
		Command:      fmt.Sprintf("cat > %s", handlerJSONFile),
		Environment:  "default",
		Organization: "default",
	}
	output, err := sensuctl.run("handler", "create", handler.Name,
		"--type", handler.Type,
		"--command", handler.Command,
		"--organization", handler.Organization,
		"--environment", handler.Environment,
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
	}
	output, err = sensuctl.run("check", "create", check.Name,
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--subscriptions", strings.Join(check.Subscriptions, ","),
		"--handlers", strings.Join(check.Handlers, ","),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	time.Sleep(10 * time.Second)

	// There should be a stored event
	output, err = sensuctl.run("event", "info", agent.ID, check.Name)
	assert.NoError(t, err, string(output))

	event := types.Event{}
	json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestEventHandler", event.Entity.ID)
	assert.Equal(t, "test", event.Check.Config.Name)

	// There should be a JSON event file in the OS temp directory
	_, err = os.Stat(handlerJSONFile)
	assert.NoError(t, err)
}
