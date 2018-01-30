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
	handling := true
	var events []*types.Event

	go func() {
		for handling {
			select {
			case msg := <-ch:
				event := &types.Event{}
				assert.NoError(json.Unmarshal(msg.Payload, event))
				assert.NotNil(event)
				assert.Equal(checkConfig.Name, event.Check.Config.Name)
				events = append(events, event)
			}
		}
	}()

	agent.handleCheck(payload)
	time.Sleep(3 * time.Second)
	agent.handleCheck(payload)
	time.Sleep(3 * time.Second)
	handling = false
	// there should be 2 events for 2 non-overlapping check executions
	assert.Equal(2, len(events))

	handling = true
	events = nil
	go func() {
		for handling {
			select {
			case msg := <-ch:
				event := &types.Event{}
				assert.NoError(json.Unmarshal(msg.Payload, event))
				assert.NotNil(event)
				events = append(events, event)
			}
		}
	}()

	agent.handleCheck(payload)
	agent.handleCheck(payload)
	time.Sleep(3 * time.Second)
	handling = false
	// there should be 1 event for 2 overlapping check executions
	assert.Equal(1, len(events))
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
