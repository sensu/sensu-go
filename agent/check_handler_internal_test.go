package agent

import (
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestExecuteCheck(t *testing.T) {
	check := types.FixtureCheck("check")

	// Does this work on windows?
	check.Command = "true"

	event := &types.Event{}
	event.Check = check

	config := NewConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	agent.executeCheck(event)

	msg := <-ch

	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 0, event.Check.Status)

	check.Command = "false"

	agent.executeCheck(event)

	msg = <-ch

	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, 1, event.Check.Status)
}
