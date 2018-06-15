// +build !windows

package agent

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEnvVars(t *testing.T) {
	checkConfig := types.FixtureCheckConfig("check")
	checkConfig.EnvVars = []string{"FOO=BAR"}
	request := &types.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true
	checkConfig.Command = "echo $FOO"

	config := FixtureConfig()
	agent := NewAgent(config)
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	agent.executeCheck(request)
	msg := <-ch
	event := &types.Event{}
	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, event.Check.Output, "BAR\n")
}
