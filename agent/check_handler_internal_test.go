package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

var binDir = filepath.Join("..", "bin")
var toolsDir = filepath.Join(binDir, "tools")

func TestHandleCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	truePath := testutil.CommandPath(filepath.Join(toolsDir, "true"))
	checkConfig.Command = truePath

	request := &types.CheckRequest{Config: checkConfig}
	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	config := FixtureConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 5)
	agent.sendq = ch

	// check is already in progress, it shouldn't execute
	agent.inProgressMu.Lock()
	agent.inProgress[request.Config.Name] = request.Config
	agent.inProgressMu.Unlock()
	assert.Error(agent.handleCheck(payload))

	// check is not in progress, it should execute
	agent.inProgressMu.Lock()
	delete(agent.inProgress, request.Config.Name)
	agent.inProgressMu.Unlock()
	assert.NoError(agent.handleCheck(payload))
}

func TestExecuteCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	request := &types.CheckRequest{Config: checkConfig}
	checkConfig.Stdin = true

	config := FixtureConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	truePath := testutil.CommandPath(filepath.Join(toolsDir, "true"))
	checkConfig.Command = truePath
	checkConfig.Timeout = 10

	agent.executeCheck(request)

	msg := <-ch

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(0), event.Check.Status)

	falsePath := testutil.CommandPath(filepath.Join(toolsDir, "false"))
	checkConfig.Command = falsePath

	agent.executeCheck(request)

	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(1), event.Check.Status)

	sleepPath := testutil.CommandPath(filepath.Join(toolsDir, "sleep"), "5")
	checkConfig.Command = sleepPath
	checkConfig.Timeout = 1

	agent.executeCheck(request)

	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(int32(2), event.Check.Status)
}

func TestPrepareCheck(t *testing.T) {
	assert := assert.New(t)

	config := FixtureConfig()
	agent := NewAgent(config)

	// Invalid check
	check := types.FixtureCheckConfig("check")
	check.Interval = 0
	assert.False(agent.prepareCheck(check))

	// Valid check
	check.Interval = 60
	assert.True(agent.prepareCheck(check))
}
