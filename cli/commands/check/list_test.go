package check

import (
	"testing"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/test"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type MockCheckList struct {
	client.RestClient
}

func (c *MockCheckList) ListChecks() ([]types.Check, error) {
	return []types.Check{
		*types.FixtureCheck("one"),
		*types.FixtureCheck("two"),
	}, nil
}

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{Client: &MockCheckList{}}
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.Run, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("checks", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	stdout := &test.StdoutCapture{}

	cli := &cli.SensuCli{Client: &MockCheckList{}}
	cmd := ListCommand(cli)

	stdout.StartCapture()
	cmd.Run(cmd, []string{})
	stdout.StopCapture()

	assert.Regexp("one.*two", stdout.Bytes)
}
