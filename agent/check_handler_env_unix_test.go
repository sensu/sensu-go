//go:build !windows
// +build !windows

package agent

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
)

func TestEnvVars(t *testing.T) {
	checkConfig := v2.FixtureCheckConfig("check")
	checkConfig.EnvVars = []string{"FOO=BAR"}
	request := &v2.CheckRequest{Config: checkConfig, Issued: time.Now().Unix()}
	checkConfig.Stdin = true
	checkConfig.Command = "echo $FOO"

	config, cleanup := FixtureConfig()
	defer cleanup()
	agent, err := NewAgent(config)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan *transport.Message, 1)
	agent.sendq = ch

	entity := agent.getAgentEntity()
	agent.executeCheck(context.Background(), request, entity)
	msg := <-ch
	event := &v2.Event{}
	assert.NoError(t, json.Unmarshal(msg.Payload, event))
	assert.NotZero(t, event.Timestamp)
	assert.Equal(t, event.Check.Output, "BAR\n")
}
