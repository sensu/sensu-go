package e2e

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/require"
)

const timeWindow = `{"days":{"all":[{"begin":"12:00 AM","end":"11:59 PM"},{"begin":"11:00 PM","end":"1:00 AM"}]}}`

func TestCheckSubdue(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestCheckScheduling",
		BackendURLs: []string{backend.WSURL},
	}
	_, cleanup = newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that publish check requests
	createCheck(t, sensuctl)

	// Make sure the check exists
	check := getCheck(t, sensuctl)
	require.NotNil(t, check)

	// Give it few seconds to make sure we're not publishing check requests.
	time.Sleep(15 * time.Second)

	checkNoEvent(t, sensuctl)
}

func createCheck(t *testing.T, ctl *sensuCtl) {
	ctl.SetStdin(strings.NewReader(timeWindow))
	defer func() {
		ctl.SetStdin(os.Stdin)
	}()
	out, err := ctl.run(
		"check", "create", "mycheck",
		"--publish",
		"--interval", "1",
		"--subscriptions", "test",
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

func checkNoEvent(t *testing.T, ctl *sensuCtl) *types.Event {
	var event types.Event

	out, err := ctl.run(
		"event", "info", "TestCheckScheduling", "mycheck",
		"--format", "json",
	)

	require.NoError(t, err, string(out))

	if len(out) == 0 || string(out) == "not found" {
		return nil
	}

	require.NoError(t, json.Unmarshal(out, &event))

	return &event
}
