package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSilencing(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend()
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestSilencing",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Create a handler that creates files within a temporary directory so we can
	// easily determine if a given event has been handled
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu-handler")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	output, err := sensuctl.run("handler", "create", "touch",
		"--organization", "default",
		"--environment", "default",
		"--timeout", "10",
		"--type", "pipe",
		"--command", fmt.Sprintf("touch %s/$(date +%%s)", tmpDir),
	)
	assert.NoError(t, err, string(output))

	// Create a dumb check that returns an error, so it gets handled, and which
	// will later be silenced and attach it to our agent via the subscription and
	// use our previously defined handler
	output, err = sensuctl.run("check", "create", "check_silencing",
		"--command", "return 2",
		"--interval", "1",
		"--subscriptions", "test",
		"--handlers", "touch",
		"--organization", "default",
		"--environment", "default",
		"--publish",
	)
	assert.NoError(t, err, string(output))

	// Wait for the agent to send a check result
	time.Sleep(10 * time.Second)

	// We should have an event for that agent and check combinaison and it should
	// not be silenced
	output, err = sensuctl.run("event", "info", agent.ID, "check_silencing")
	assert.NoError(t, err, string(output))
	event := types.Event{}
	json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.Empty(t, event.Silenced)

	// Create a silencing entry for that particular entity and check
	output, err = sensuctl.run("silenced", "create",
		"--subscription", "entity:TestSilencing",
		"--check", "check_silencing",
	)
	assert.NoError(t, err, string(output))

	// Retrieve the number of files created by our handler, representing the number
	// of times our handler was ran
	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		log.Fatal(err)
	}
	count := len(files)
	fmt.Println(count)
}
