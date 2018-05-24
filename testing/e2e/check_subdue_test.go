package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const timeWindow = `{"days":{"all":[{"begin":"12:00AM UTC","end":"11:59PM UTC"},` +
	`{"begin":"11:00PM UTC","end":"1:00AM UTC"}]}}`

func TestCheckSubdue(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	ctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestCheckSubdue",
	}
	_, cleanup = newAgent(agentConfig, ctl, t)
	defer cleanup()

	// Create a check that publish check requests
	createCheck(t, ctl)

	// Make sure the check exists
	check := getCheck(t, ctl)
	require.NotNil(t, check)

	getCheckEventCmd := []string{"event", "info", "TestCheckSubdue", "mycheck",
		"--format", "json",
		"--organization", ctl.Organization,
		"--environment", ctl.Environment,
	}

	var output []byte
	var err error
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = ctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event received: %s", string(output))
	}

	event1 := &types.Event{}
	require.NoError(t, json.Unmarshal(output, event1))
	require.NotNil(t, event1)

	// Subdue the check
	subdueCheck(t, ctl, timeWindow)

	// FIXME: Give it a few seconds to pick up the change
	time.Sleep(10 * time.Second)

	event2 := getEvent(t, ctl)

	// FIXME: wait *again* to make sure check requests are not being published anymore
	time.Sleep(10 * time.Second)
	event3 := getEvent(t, ctl)

	assert.Equal(t, event2.Check.History, event3.Check.History)

	// Un-subdue the check
	subdueCheck(t, ctl, "{}")

	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = ctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event4 := &types.Event{}
		if err := json.Unmarshal(output, event4); err != nil || event4 == nil {
			return false, nil
		}

		// Ensure we received a new keepalive message from the agent
		if event3.Timestamp == event4.Timestamp {
			// Let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no new event received since check was un-subdue %s", string(output))
	}
}

func createCheck(t *testing.T, ctl *sensuCtl) {
	out, err := ctl.run(
		"check", "create", "mycheck",
		"--publish",
		"--interval", "1",
		"--subscriptions", "test",
		"--command", "true",
		"--organization", ctl.Organization,
		"--environment", ctl.Environment,
	)
	require.NoError(t, err, string(out))
}

func getCheck(t *testing.T, ctl *sensuCtl) *types.Check {
	var check types.Check

	out, err := ctl.run(
		"check", "info", "mycheck",
		"--format", "json",
		"--organization", ctl.Organization,
		"--environment", ctl.Environment,
	)

	require.NoError(t, err, string(out))

	if len(out) == 0 || string(out) == "not found" {
		return nil
	}

	require.NoError(t, json.Unmarshal(out, &check))

	return &check
}

func getEvent(t *testing.T, ctl *sensuCtl) *types.Event {
	var event types.Event

	out, err := ctl.run(
		"event", "info", "TestCheckSubdue", "mycheck",
		"--format", "json",
		"--organization", ctl.Organization,
		"--environment", ctl.Environment,
	)

	require.NoError(t, err, string(out))

	if len(out) == 0 || string(out) == "not found" {
		return nil
	}

	require.NoError(t, json.Unmarshal(out, &event))

	return &event
}

func subdueCheck(t *testing.T, ctl *sensuCtl, data string) {
	ctl.SetStdin(strings.NewReader(data))
	defer func() {
		ctl.SetStdin(os.Stdin)
	}()
	_, err := ctl.run("check", "set-subdue", "mycheck",
		"--organization", ctl.Organization,
		"--environment", ctl.Environment,
	)
	require.NoError(t, err)
}
