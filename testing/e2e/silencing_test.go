// +build !windows

package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TestSilencing ensures that the silencing functionality works as expected, by
// testing a happy path scenario where we have a handler that produces files
// into a temporary directory when handling events. So we simply make sure that
// this handler does not run when the associated check and entity are silenced,
// and therefore that the event is not passing through pipelined, by counting
// the number of files created.
func TestSilencing(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestSilencing",
		BackendURLs: []string{backend.WSURL},
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a handler that creates files within a temporary directory so we can
	// easily determine if a given event has been handled
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu-handler")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}()

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
		"--command", "false",
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
	_ = json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.Empty(t, event.Silenced)

	// Create a silencing entry for that particular entity and check
	output, err = sensuctl.run("silenced", "create",
		"--subscription", "entity:TestSilencing",
		"--check", "check_silencing",
		"--reason", "to test silencing",
	)
	assert.NoError(t, err, string(output))

	// Wait for new check results so the event gets updated with this new
	// silenced entry
	time.Sleep(2 * time.Second)

	// Retrieve the number of files created by our handler, representing the
	// number of times our handler was ran. Since the check is silenced, this
	// number should not move from this point
	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	count1 := len(files)

	// Make sure the event is marked as silenced now
	output, err = sensuctl.run("event", "info", agent.ID, "check_silencing")
	assert.NoError(t, err, string(output))
	event = types.Event{}
	_ = json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.Silenced)

	// Make sure the keepalive event is not silenced by our silenced entry
	// N.B. This is currently broken, see
	// https://github.com/sensu/sensu-go/issues/707
	// output, err = sensuctl.run("event", "info", agent.ID, "keepalive")
	// assert.NoError(t, err, string(output))
	// event = types.Event{}
	// json.Unmarshal(output, &event)
	// assert.NotNil(t, event)
	// assert.Empty(t, event.Silenced)

	// The number of files created by the handler should not have increased since
	// the silenced entry was created
	files, err = ioutil.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	count2 := len(files)
	assert.Equal(t, count1, count2)

	// Delete the silenced entry so events are once again handled
	output, err = sensuctl.run(
		"silenced",
		"delete",
		fmt.Sprintf("entity:%s:check_silencing", agent.ID),
		"--skip-confirm",
	)
	assert.NoError(t, err, string(output))

	// Wait for new check results so the event gets updated
	time.Sleep(2 * time.Second)

	// The number of files created by the handler should have increased
	files, err = ioutil.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	count3 := len(files)
	assert.Condition(
		t,
		assert.Comparison(func() bool {
			return count3 > count2
		}),
		"the 'touch' handler did not created new files as expected",
	)

}
