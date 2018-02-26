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

const timeWindow = `{"days":{"all":[{"begin":"12:00 AM","end":"11:59 PM"},{"begin":"11:00 PM","end":"1:00 AM"}]}}`

func TestCheckSubdue(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	ctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestCheckSubdue",
		BackendURLs: []string{backend.WSURL},
	}
	_, cleanup = newAgent(agentConfig, ctl, t)
	defer cleanup()

	// Create a check that publish check requests
	createCheck(t, ctl)

	// Make sure the check exists
	check := getCheck(t, ctl)
	require.NotNil(t, check)

	// FIXME: Give it few seconds to make sure we're not publishing check requests.
	time.Sleep(15 * time.Second)

	event1 := getEvent(t, ctl)
	require.NotNil(t, event1)

	if len(event1.Check.History) == 0 {
		t.Error("missing check history")
	}

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

	// FIXME: Give it a few seconds to pick up the change
	time.Sleep(15 * time.Second)

	event4 := getEvent(t, ctl)
	if len(event4.Check.History) <= len(event3.Check.History) {
		t.Error("check did not start executing again")
	}
}

func createCheck(t *testing.T, ctl *sensuCtl) {
	out, err := ctl.run(
		"check", "create", "mycheck",
		"--publish",
		"--interval", "1",
		"--subscriptions", "test",
		"--command", "true",
	)
	require.NoError(t, err, string(out))
}

func getCheck(t *testing.T, ctl *sensuCtl) *types.Check {
	var check types.Check

	out, err := ctl.run(
		"check", "info", "mycheck",
		"--format", "json",
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
	_, err := ctl.run("check", "set-subdue", "mycheck")
	require.NoError(t, err)
}
