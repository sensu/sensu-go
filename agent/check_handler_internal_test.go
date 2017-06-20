package agent

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

var binDir = filepath.Join("..", "bin")
var toolsDir = filepath.Join(binDir, "tools")

func TestExecuteCheck(t *testing.T) {
	checkConfig := types.FixtureCheckConfig("check")
	request := &types.CheckRequest{Config: checkConfig}

	config := NewConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	truePath := util.CommandPath(filepath.Join(toolsDir, "true"))
	checkConfig.Command = truePath

	agent.executeCheck(request)

	msg := <-ch

	event := &types.Event{}
	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 0, event.Check.Status)

	falsePath := util.CommandPath(filepath.Join(toolsDir, "false"))
	checkConfig.Command = falsePath

	agent.executeCheck(request)

	msg = <-ch

	event = &types.Event{}
	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 1, event.Check.Status)
}
