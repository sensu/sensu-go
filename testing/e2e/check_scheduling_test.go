package e2e

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckScheduling(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestCheckScheduling",
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that publish check requests
	check := types.FixtureCheckConfig("TestCheckScheduling")
	check.Publish = true
	check.Interval = 1
	check.Subscriptions = []string{"test"}
	check.Organization = sensuctl.Organization
	check.Environment = sensuctl.Environment

	getCheckEventCmd := []string{"event", "info", agent.ID, check.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	}

	out, err := sensuctl.run("check", "create", check.Name,
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--runtime-assets", strings.Join(check.RuntimeAssets, ","),
		"--subscriptions", strings.Join(check.Subscriptions, ","),
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)

	require.NoError(t, err, string(out))

	out, err = sensuctl.run("check", "info", check.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)

	require.NoError(t, err, string(out))

	// Make sure we have at least an event for this check
	var output []byte
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event received: %s", string(output))
	}

	// Stop publishing check requests
	out, err = sensuctl.run("check", "set-publish", check.Name, "false",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)

	require.NoError(t, err, string(out))

	out, err = sensuctl.run("check", "info", check.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)
	require.NoError(t, err, string(out))

	// Give it few seconds to make sure we are not publishing check requests
	time.Sleep(20 * time.Second)

	// Retrieve the number of check results sent
	out, err = sensuctl.run("event", "info", agent.ID, check.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)
	require.NoError(t, err, string(out))
	var event types.Event
	if err := json.Unmarshal(out, &event); err != nil {
		t.Fatal(err)
	}

	count1 := len(event.Check.History)

	// Give it few seconds to make sure we did not published additional check requests
	time.Sleep(10 * time.Second)

	// Retrieve (again) the number of check results sent
	out, err = sensuctl.run("event", "info", agent.ID, check.Name,
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)
	require.NoError(t, err, string(out))
	if err := json.Unmarshal(out, &event); err != nil {
		t.Fatal(err)
	}

	count2 := len(event.Check.History)

	// Make sure no new check results were sent
	assert.Equal(t, count1, count2)

	// Start publishing check requests again
	out, err = sensuctl.run("check", "set-publish", check.Name, "true",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)
	require.NoError(t, err, string(out))

	var count3 int
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}
		count3 = len(event.Check.History)

		if count2 == count3 {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no new event received since publish set to true again: %s", string(output))
	}

	// Change the check schedule to cron
	out, err = sensuctl.run("check", "set-cron", check.Name, "* * * * *",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment)
	require.NoError(t, err, string(out))

	// Since the cronjob runs every minute, we'll need a longer backoff strategy
	longBackoff := retry.ExponentialBackoff{
		InitialDelayInterval: 10 * time.Second,
		MaxDelayInterval:     90 * time.Second,
		MaxRetryAttempts:     0, // Unlimited attempts
		Multiplier:           1.1,
	}

	var count4 int
	if err := longBackoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run(getCheckEventCmd...); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}
		count4 = len(event.Check.History)

		if count3 == count4 {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no event received since schedule was set to cron: %s", string(output))
	}
}
