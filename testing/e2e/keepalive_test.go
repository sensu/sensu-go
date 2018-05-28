package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type eventsTest struct {
	cleanup  func()
	ap       *agentProcess
	sensuctl *sensuCtl
}

func newEventsTest(t *testing.T) *eventsTest {
	t.Parallel()
	test := &eventsTest{}

	// Initializes sensuctl
	sensuctl, sensuctlCleanup := newSensuCtl(t)

	// Start the agent
	agentConfig := agentConfig{
		ID: "TestKeepalives",
	}
	agent, agentCleanup := newAgent(agentConfig, sensuctl, t)

	test.ap = agent
	test.sensuctl = sensuctl

	test.cleanup = func() {
		agentCleanup()
		sensuctlCleanup()
	}

	// Allow time agent connection to be established, etcd to start,
	// keepalive to be sent, etc.
	var output []byte
	var err error
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info",
			agent.ID, "keepalive"); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no keepalive received: %s", string(output))
	}

	return test
}

func TestKeepaliveEvent(t *testing.T) {
	test := newEventsTest(t)
	defer test.cleanup()

	assert := assert.New(t)

	output, err := test.sensuctl.run(
		"event", "list",
		"--organization", test.sensuctl.Organization,
		"--environment", test.sensuctl.Environment,
	)
	assert.NoError(err)

	events := []types.Event{}
	assert.NoError(json.Unmarshal(output, &events))

	assert.NotZero(len(events))

	seen := false
	for _, ev := range events {
		if ev.Check.Name == "keepalive" {
			seen = true
			assert.Equal("TestKeepalives", ev.Entity.ID)
			assert.NotZero(ev.Timestamp)
			assert.Equal("passing", ev.Check.State)
		}
	}
	assert.True(seen)
}

func TestEntity(t *testing.T) {
	test := newEventsTest(t)
	defer test.cleanup()
	assert := assert.New(t)

	// Retrieve the entitites
	output, err := test.sensuctl.run(
		"entity", "list",
		"--organization", test.sensuctl.Organization,
		"--environment", test.sensuctl.Environment,
	)
	require.NoError(t, err)

	entities := []types.Entity{}
	assert.NoError(json.Unmarshal(output, &entities))

	assert.Equal(1, len(entities))
	assert.Equal("TestKeepalives", entities[0].ID)
	assert.Equal("agent", entities[0].Class)
	assert.NotEmpty(entities[0].System.Hostname)
	assert.NotZero(entities[0].LastSeen)
}

func TestCheck(t *testing.T) {
	test := newEventsTest(t)
	defer test.cleanup()
	assert := assert.New(t)

	falsePath := testutil.CommandPath(filepath.Join(binDir, "false"))
	falseAbsPath, err := filepath.Abs(falsePath)
	assert.NoError(err)
	assert.NotEmpty(falseAbsPath)

	// Create a standard check
	checkName := "test_check"
	_, err = test.sensuctl.run("check", "create", checkName,
		"--command", falseAbsPath,
		"--interval", "1",
		"--subscriptions", "test",
		"--publish",
		"--organization", test.sensuctl.Organization,
		"--environment", test.sensuctl.Environment,
	)
	assert.NoError(err)

	// Make sure the check has been properly created
	output, err := test.sensuctl.run(
		"check", "info", checkName,
		"--organization", test.sensuctl.Organization,
		"--environment", test.sensuctl.Environment,
	)
	require.NoError(t, err)

	result := types.CheckConfig{}
	assert.NoError(json.Unmarshal(output, &result))
	assert.Equal(result.Name, checkName)

	// Allow enough time for the check to run.
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = test.sensuctl.run("event", "info", test.ap.ID, checkName,
			"--organization", test.sensuctl.Organization,
			"--environment", test.sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no keepalive received: %s", string(output))
	}
}

func TestHTTPAPI(t *testing.T) {
	test := newEventsTest(t)
	defer test.cleanup()
	assert := assert.New(t)

	newEvent := types.FixtureEvent(test.ap.ID, "proxy-check")
	newEvent.Check.Organization = test.sensuctl.Organization
	newEvent.Check.Environment = test.sensuctl.Environment
	encoded, _ := json.Marshal(newEvent)
	url := fmt.Sprintf("http://127.0.0.1:%d/events", test.ap.APIPort)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(encoded))

	client := &http.Client{}
	res, err := client.Do(req)
	require.NoError(t, err)
	defer func() {
		assert.NoError(res.Body.Close())
	}()

	// Give it a second to receive the new event
	var output []byte
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = test.sensuctl.run("event", "info", test.ap.ID, "proxy-check",
			"--organization", test.sensuctl.Organization,
			"--environment", test.sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no events received: %s", string(output))
	}
}

func TestKeepaliveTimeout(t *testing.T) {
	// Initializes sensuctl
	sensuctl, sensuctlCleanup := newSensuCtl(t)

	// Start the agent
	agentConfig := agentConfig{
		ID:               "TestKeepalives",
		KeepaliveTimeout: 5,
	}
	agent, agentCleanup := newAgent(agentConfig, sensuctl, t)

	defer func() {
		agentCleanup()
		sensuctlCleanup()
	}()

	// Allow time for agent connection to be established, etcd to start,
	// keepalive to be sent, etc.
	var output []byte
	var err error
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", agent.ID, "keepalive",
			"--organization", sensuctl.Organization,
			"--environment", sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no keepalive received: %s", string(output))
	}

	event := &types.Event{}
	assert.NoError(t, json.Unmarshal(output, event))

	assert.NotNil(t, event)
	assert.Equal(t, "TestKeepalives", event.Entity.ID)
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, "passing", event.Check.State)
	assert.Equal(t, uint32(0), event.Check.Status)

	// Stop the agent, and restart with a new KeepaliveTimeout
	assert.NoError(t, agent.Terminate())

	agentConfig.KeepaliveTimeout = 1
	agentConfig.KeepaliveInterval = 2
	agent, agentCleanup = newAgent(agentConfig, sensuctl, t)

	// Allow time for agent connection to be established, etcd to start,
	// keepalive to be sent, etc.
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", agent.ID, "keepalive",
			"--organization", sensuctl.Organization,
			"--environment", sensuctl.Environment); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		if event.Timestamp == 0 || event.Check.State != "failing" || event.Check.Status != uint32(1) {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no failing keepalive received: %s", string(output))
	}
}
