package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type eventsTest struct {
	bep      *backendProcess
	cleanup  func()
	ap       *agentProcess
	sensuctl *sensuCtl
}

func newEventsTest(t *testing.T) *eventsTest {
	test := &eventsTest{}

	// Start the backend
	backend, backendCleanup := newBackend(t)

	// Initializes sensuctl
	sensuctl, sensuctlCleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestKeepalives",
		BackendURLs: []string{backend.WSURL},
	}
	agent, agentCleanup := newAgent(agentConfig, sensuctl, t)

	test.ap = agent
	test.bep = backend
	test.sensuctl = sensuctl

	test.cleanup = func() {
		backendCleanup()
		agentCleanup()
		sensuctlCleanup()
	}

	// Allow time agent connection to be established, etcd to start,
	// keepalive to be sent, etc.
	time.Sleep(10 * time.Second)

	return test
}

func TestKeepaliveEvent(t *testing.T) {
	test := newEventsTest(t)
	defer test.cleanup()
	t.Parallel()

	assert := assert.New(t)

	output, err := test.sensuctl.run("event", "list")
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
	t.Parallel()

	// Retrieve the entitites
	output, err := test.sensuctl.run("entity", "list")
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
	t.Parallel()

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
	)
	assert.NoError(err)

	// Make sure the check has been properly created
	output, err := test.sensuctl.run("check", "info", checkName)
	require.NoError(t, err)

	result := types.CheckConfig{}
	assert.NoError(json.Unmarshal(output, &result))
	assert.Equal(result.Name, checkName)

	// Allow enough time for the check to run.
	time.Sleep(20 * time.Second)
	output, err = test.sensuctl.run("event", "info", test.ap.ID, checkName)
	require.NoError(t, err)

	event := types.Event{}
	assert.NoError(json.Unmarshal(output, &event))
	assert.NotNil(event)
	assert.NotNil(event.Check)
	assert.NotNil(event.Entity)
	assert.Equal("TestKeepalives", event.Entity.ID)
	assert.Equal(checkName, event.Check.Name)
}

func TestHTTPAPI(t *testing.T) {
	test := newEventsTest(t)
	defer test.cleanup()
	assert := assert.New(t)
	t.Parallel()

	newEvent := types.FixtureEvent(test.ap.ID, "proxy-check")
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
	time.Sleep(5 * time.Second)

	// Make sure the new event has been received
	output, err := test.sensuctl.run("event", "info", test.ap.ID, "proxy-check")
	assert.NoError(err, string(output))
	assert.NotNil(output)
}

func TestKeepaliveTimeout(t *testing.T) {
	// Start the backend
	backend, backendCleanup := newBackend(t)

	// Initializes sensuctl
	sensuctl, sensuctlCleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")

	// Start the agent
	agentConfig := agentConfig{
		ID:               "TestKeepalives",
		BackendURLs:      []string{backend.WSURL},
		KeepaliveTimeout: 5,
	}
	agent, agentCleanup := newAgent(agentConfig, sensuctl, t)

	defer func() {
		backendCleanup()
		agentCleanup()
		sensuctlCleanup()
	}()

	// Allow time for agent connection to be established, etcd to start,
	// keepalive to be sent, etc.
	time.Sleep(10 * time.Second)

	output, err := sensuctl.run("event", "info", agent.ID, "keepalive")
	assert.NoError(t, err)

	event := types.Event{}
	assert.NoError(t, json.Unmarshal(output, &event))

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
	time.Sleep(10 * time.Second)

	output, err = sensuctl.run("event", "info", agent.ID, "keepalive")
	assert.NoError(t, err)

	event = types.Event{}
	assert.NoError(t, json.Unmarshal(output, &event))

	assert.NotNil(t, event)
	assert.Equal(t, "TestKeepalives", event.Entity.ID)
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, "failing", event.Check.State)
	assert.Equal(t, uint32(1), event.Check.Status)
}
