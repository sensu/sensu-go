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
	"github.com/stretchr/testify/suite"
)

type EventTestSuite struct {
	suite.Suite
	bep      *backendProcess
	cleanup  func()
	ap       *agentProcess
	sensuctl *sensuCtl
}

func (suite *EventTestSuite) SetupSuite() {
	// Start the backend
	backend, backendCleanup := newBackend(suite.T())

	// Initializes sensuctl
	sensuctl, sensuctlCleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestKeepalives",
		BackendURLs: []string{backend.WSURL},
	}
	agent, agentCleanup := newAgent(agentConfig, sensuctl, suite.T())

	suite.ap = agent
	suite.bep = backend
	suite.sensuctl = sensuctl
	suite.cleanup = func() {
		backendCleanup()
		agentCleanup()
		sensuctlCleanup()
	}

	// Allow time agent connection to be established, etcd to start,
	// keepalive to be sent, etc.
	time.Sleep(10 * time.Second)
}

func (suite *EventTestSuite) TearDownSuite() {
	suite.cleanup()
}

func (suite *EventTestSuite) TestKeepaliveEvent() {
	assert := suite.Assert()

	output, err := suite.sensuctl.run("event", "list")
	assert.NoError(err)

	events := []types.Event{}
	suite.NoError(json.Unmarshal(output, &events))

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

func (suite *EventTestSuite) TestEntity() {
	assert := suite.Assert()

	// Retrieve the entitites
	output, err := suite.sensuctl.run("entity", "list")
	assert.NoError(err)

	entities := []types.Entity{}
	suite.NoError(json.Unmarshal(output, &entities))

	assert.Equal(1, len(entities))
	assert.Equal("TestKeepalives", entities[0].ID)
	assert.Equal("agent", entities[0].Class)
	assert.NotEmpty(entities[0].System.Hostname)
	assert.NotZero(entities[0].LastSeen)
}

func (suite *EventTestSuite) TestCheck() {
	assert := suite.Assert()

	falsePath := testutil.CommandPath(filepath.Join(binDir, "false"))
	falseAbsPath, err := filepath.Abs(falsePath)
	assert.NoError(err)
	assert.NotEmpty(falseAbsPath)

	// Create a standard check
	checkName := "test_check"
	_, err = suite.sensuctl.run("check", "create", checkName,
		"--command", falseAbsPath,
		"--interval", "1",
		"--subscriptions", "test",
		"--publish",
	)
	assert.NoError(err)

	// Make sure the check has been properly created
	output, err := suite.sensuctl.run("check", "info", checkName)
	assert.NoError(err)

	result := types.CheckConfig{}
	suite.NoError(json.Unmarshal(output, &result))
	assert.Equal(result.Name, checkName)

	// Allow enough time for the check to run.
	time.Sleep(20 * time.Second)
	output, err = suite.sensuctl.run("event", "info", suite.ap.ID, checkName)
	assert.NoError(err)

	event := types.Event{}
	suite.NoError(json.Unmarshal(output, &event))
	assert.NotNil(event)
	assert.NotNil(event.Check)
	assert.NotNil(event.Entity)
	assert.Equal("TestKeepalives", event.Entity.ID)
	assert.Equal(checkName, event.Check.Name)
}

func (suite *EventTestSuite) TestHTTPAPI() {
	assert := suite.Assert()

	newEvent := types.FixtureEvent(suite.ap.ID, "proxy-check")
	encoded, _ := json.Marshal(newEvent)
	url := fmt.Sprintf("http://127.0.0.1:%d/events", suite.ap.APIPort)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(encoded))

	client := &http.Client{}
	res, err := client.Do(req)
	assert.NoError(err)
	defer func() {
		suite.NoError(res.Body.Close())
	}()

	// Give it a second to receive the new event
	time.Sleep(5 * time.Second)

	// Make sure the new event has been received
	output, err := suite.sensuctl.run("event", "info", suite.ap.ID, "proxy-check")
	assert.NoError(err, string(output))
	assert.NotNil(output)
}

func TestEventTestSuite(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(EventTestSuite))
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
	assert.Equal(t, int32(0), event.Check.Status)

	// Kill the agent, and restart with a new KeepaliveTimeout
	assert.NoError(t, agent.Kill())

	// Allow time agent connection to be killed
	time.Sleep(10 * time.Second)

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
	assert.Equal(t, int32(1), event.Check.Status)
}
