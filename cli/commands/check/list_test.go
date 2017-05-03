package check

import (
	"errors"
	"os"
	"testing"

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

	cli := test.SimpleSensuCLI(&MockCheckList{})
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("checks", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	cli := test.SimpleSensuCLI(&MockCheckList{})
	cmd := ListCommand(cli)

	stdout.Start()
	cmd.RunE(cmd, []string{})
	stdout.Stop()

	assert.NotEmpty(stdout.Output())
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	cli := test.SimpleSensuCLI(&MockCheckList{})
	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "table")

	stdout.Start()
	cmd.RunE(cmd, []string{})
	stdout.Stop()

	assert.NotEmpty(stdout.Output())
}

var errorFetchingChecks = errors.New("500 err")

type MockCheckListErr struct {
	client.RestClient
}

func (c *MockCheckListErr) ListChecks() ([]types.Check, error) {
	return nil, errorFetchingChecks
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.SimpleSensuCLI(&MockCheckListErr{})
	cmd := ListCommand(cli)

	err := cmd.RunE(cmd, []string{})
	assert.NotNil(err)
	assert.Equal(errorFetchingChecks, err)
}
