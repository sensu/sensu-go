package hook

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("info", cmd.Use)
	assert.Regexp("hook", cmd.Short)
}

func TestInfoCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchHook", "in").Return(corev2.FixtureHookConfig("name-one"), nil)

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
	assert.Nil(err)
}

func TestInfoCommandRunMissingArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Error(err)
	assert.NotEmpty(out)
	assert.Contains(out, "Usage")
}

func TestInfoCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchHook", "in").Return(corev2.FixtureHookConfig("name-one"), nil)

	cmd := InfoCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotEmpty(out)
	assert.Contains(out, "Name")
	assert.Contains(out, "Timeout")
	assert.Contains(out, "Command")
	assert.Nil(err)
}

func TestInfoCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchHook", "in").Return(&corev2.HookConfig{}, errors.New("my-err"))

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}
