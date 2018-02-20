package hook

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ShowCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("info", cmd.Use)
	assert.Regexp("hook", cmd.Short)
}

func TestShowCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchHook", "in").Return(types.FixtureHookConfig("name-one"), nil)

	cmd := ShowCommand(cli)
	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
	assert.Nil(err)
}

func TestShowCommandRunMissingArgs(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ShowCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Error(err)
	assert.NotEmpty(out)
	assert.Contains(out, "Usage")
}

func TestShowCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchHook", "in").Return(types.FixtureHookConfig("name-one"), nil)

	cmd := ShowCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotEmpty(out)
	assert.Contains(out, "Name")
	assert.Contains(out, "Timeout")
	assert.Contains(out, "Command")
	assert.Nil(err)
}

func TestShowCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchHook", "in").Return(&types.HookConfig{}, errors.New("my-err"))

	cmd := ShowCommand(cli)
	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}
