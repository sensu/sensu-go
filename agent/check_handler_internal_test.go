// +build integration

package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

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
	truePath := testutil.CommandPath(filepath.Join(toolsDir, "sleep 2"))
	checkConfig.Command = truePath

	request := &types.CheckRequest{Config: checkConfig}
	payload, err := json.Marshal(request)
	if err != nil {
		assert.FailNow("error marshaling check request")
	}

	config := NewConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 5)
	agent.sendq = ch

	assert.NoError(agent.handleCheck(payload))
	assert.Error(agent.handleCheck(payload))
	assert.NotNil(t, <-agent.sendq)
	time.Sleep(3 * time.Second)
	select {
	case msg := <-agent.sendq:
		assert.FailNow("received unexpected message: %s", msg)
	default:
	}

	assert.NoError(agent.handleCheck(payload))
	time.Sleep(4 * time.Second)
	assert.NoError(agent.handleCheck(payload))
	assert.NotNil(t, <-agent.sendq)
	select {
	case msg := <-agent.sendq:
		assert.FailNow("received unexpected message: %s", msg)
	default:
	}
}

func TestExecuteCheck(t *testing.T) {
	assert := assert.New(t)

	checkConfig := types.FixtureCheckConfig("check")
	request := &types.CheckRequest{Config: checkConfig}
	checkConfig.Stdin = true

	config := NewConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	truePath := testutil.CommandPath(filepath.Join(toolsDir, "true"))
	checkConfig.Command = truePath

	agent.executeCheck(request)

	msg := <-ch

	event := &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(0, event.Check.Status)

	falsePath := testutil.CommandPath(filepath.Join(toolsDir, "false"))
	checkConfig.Command = falsePath

	agent.executeCheck(request)

	msg = <-ch

	event = &types.Event{}
	assert.NoError(json.Unmarshal(msg.Payload, event))
	assert.NotZero(event.Timestamp)
	assert.EqualValues(1, event.Check.Status)
}

func TestPrepareCheck(t *testing.T) {
	assert := assert.New(t)

	config := NewConfig()
	agent := NewAgent(config)

	// Invalid check
	check := types.FixtureCheckConfig("check")
	check.Interval = 0
	assert.False(agent.prepareCheck(check))

	// Valid check
	check.Interval = 60
	assert.True(agent.prepareCheck(check))
}
