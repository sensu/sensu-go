package handler

import (
	"errors"
	"os"
	"testing"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/test"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type MockHandlerList struct {
	client.RestClient
}

func (c *MockHandlerList) ListHandlers() ([]types.Handler, error) {
	return []types.Handler{
		*types.FixtureHandler("one"),
		*types.FixtureHandler("two"),
	}, nil
}

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.SimpleSensuCLI(&MockHandlerList{})
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("handler", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	cli := test.SimpleSensuCLI(&MockHandlerList{})
	cmd := ListCommand(cli)

	stdout.Start()
	cmd.RunE(cmd, []string{})
	stdout.Stop()

	assert.NotEmpty(stdout.Output())
}

var errorFetchingHandlers = errors.New("500 err")

type MockHandlerListErr struct {
	client.RestClient
}

func (c *MockHandlerListErr) ListHandlers() ([]types.Handler, error) {
	return nil, errorFetchingHandlers
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.SimpleSensuCLI(&MockHandlerListErr{})
	cmd := ListCommand(cli)

	err := cmd.RunE(cmd, []string{})
	assert.NotNil(err)
	assert.Equal(errorFetchingHandlers, err)
}
