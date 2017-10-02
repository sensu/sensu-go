package e2e

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
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
		bep.Kill()
		ap.Kill()
	}()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	// Create an authenticated HTTP Sensu client
	clientConfig := &basic.Config{
		Cluster: basic.Cluster{
			APIUrl: backendHTTPURL,
		},
	}
	sensuClient := client.New(clientConfig)
	tokens, _ := sensuClient.CreateAccessToken(backendHTTPURL, "admin", "P@ssw0rd!")
	clientConfig.Cluster.Tokens = tokens

	handlerJSONFile := fmt.Sprintf("%s/TestEventHandler%v", os.TempDir(), os.Getpid())

	// Create a handler
	handler := &types.Handler{
		Name:         "test",
		Type:         "pipe",
		Command:      fmt.Sprintf("cat > %s", handlerJSONFile),
		Environment:  "default",
		Organization: "default",
	}
	err = sensuClient.CreateHandler(handler)
	assert.NoError(t, err)

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
	err = sensuClient.CreateCheck(check)
	assert.NoError(t, err)

	time.Sleep(10 * time.Second)

	// There should be a stored event
	event, err := sensuClient.FetchEvent(ap.AgentID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestEventHandler", event.Entity.ID)
	assert.Equal(t, "test", event.Check.Config.Name)

	// There should be a JSON event file in the OS temp directory
	_, err = os.Stat(handlerJSONFile)
	assert.NoError(t, err)
}
