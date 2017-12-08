package e2e

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test check event creation -> event handler.
func TestEventHandler(t *testing.T) {
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
		AgentID:     "TestEventHandler",
	}

	err = ap.Start()
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, bep.Kill())
		assert.NoError(t, ap.Kill())
	}()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backendHTTPURL, "default", "default", "admin", "P@ssw0rd!")
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
	output, err = sensuctl.run("event", "info", ap.AgentID, check.Name)
	assert.NoError(t, err, string(output))

	event := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestEventHandler", event.Entity.ID)
	assert.Equal(t, "test", event.Check.Config.Name)

	// There should be a JSON event file in the OS temp directory
	_, err = os.Stat(handlerJSONFile)
	assert.NoError(t, err)
}
