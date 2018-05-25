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

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestSilencing",
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
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
		"--timeout", "10",
		"--type", "pipe",
		"--command", fmt.Sprintf("touch %s/$(date +%%s)", tmpDir),
		"--filters", "not_silenced",
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
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	getCheckEventCmd := []string{"event", "info", agent.ID, "check_silencing",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	}

	// Wait for the agent to send a check result
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			// Let's retry
			return false, nil
		}

		// Ensure we received a new keepalive message from the agent
		if event.Check == nil || len(event.Check.Silenced) != 0 {
			// Let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no new keepalive received: %s", string(output))
	}

	// Create a silencing entry for that particular entity and check
	output, err = sensuctl.run("silenced", "create",
		"--subscription", "entity:TestSilencing",
		"--check", "check_silencing",
		"--reason", "to test silencing",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	assert.NoError(t, err, string(output))

	// Wait for new check results so the event gets updated with this new
	// silenced entry
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		// Ensure we received a new keepalive message from the agent
		if event.Check == nil || !event.IsSilenced() {
			// Let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event received since the check was silenced: %s", string(output))
	}

	// Retrieve the number of files created by our handler, representing the
	// number of times our handler was ran. Since the check is silenced, this
	// number should not move from this point
	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	count1 := len(files)

	// Make sure the keepalive event is not silenced by our silenced entry
	output, err = sensuctl.run(
		"event", "info", agent.ID, "keepalive",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	assert.NoError(t, err, string(output))

	event := types.Event{}
	_ = json.Unmarshal(output, &event)
	assert.NotNil(t, event)
	assert.Empty(t, event.Check.Silenced)

	// The number of files created by the handler should not have increased since
	// the silenced entry was created
	time.Sleep(5 * time.Second)
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
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	assert.NoError(t, err, string(output))

	// Wait for new check results so the event gets updated
	if err := backoff.Retry(func(retry int) (bool, error) {
		// The number of files created by the handler should have increased
		if files, err := ioutil.ReadDir(tmpDir); err != nil || len(files) == count2 {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("the 'touch' handler did not created new files as expected")
	}
}
