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
	check := types.FixtureCheck("check")

	event := &types.Event{}
	event.Check = check

	config := NewConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	truePath := util.CommandPath(filepath.Join(toolsDir, "true"))
	check.Command = truePath

	agent.executeCheck(event)

	msg := <-ch

	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 0, event.Check.Status)

	falsePath := util.CommandPath(filepath.Join(toolsDir, "false"))
	check.Command = falsePath

	agent.executeCheck(event)

	msg = <-ch

	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 1, event.Check.Status)
}
