package check

import (
	"os"
	"testing"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/test"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type MockCheckCreate struct {
	client.RestClient
}

func (m *MockCheckCreate) CreateCheck(c *types.Check) error {
	return nil
}

func TestCreateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{Client: &MockCheckCreate{}}
	cmd := CreateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("create", cmd.Use)
	assert.Regexp("checks", cmd.Short)
}

func TestCreateCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)
	stderr := test.NewFileCapture(&os.Stderr)

	cli := &cli.SensuCli{Client: &MockCheckCreate{}}
	cmd := CreateCommand(cli)

	stdout.Start()
	stderr.Start()
	err := cmd.RunE(cmd, []string{"echo 'sensu'"})
	stderr.Stop()
	stdout.Stop()

	assert.Empty(stdout.Output())
	assert.Empty(stderr.Output())
	assert.NotNil(err)
}
