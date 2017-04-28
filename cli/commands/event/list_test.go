package event

import (
	"os"
	"testing"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/test"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type MockEventList struct {
	client.RestClient
}

func (c *MockEventList) ListEvents() ([]types.Event, error) {
	return []types.Event{
		*types.FixtureEvent("1", "something"),
		*types.FixtureEvent("2", "funny"),
	}, nil
}

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{Client: &MockEventList{}}
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.Run, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("events", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	cli := &cli.SensuCli{Client: &MockEventList{}}
	cmd := ListCommand(cli)

	stdout.Start()
	cmd.Run(cmd, []string{})
	stdout.Stop()

	assert.NotEmpty(stdout.Output())
}
