package check

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
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
	assert.Regexp("check", cmd.Short)
}

func TestInfoCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchCheck", "in").Return(types.FixtureCheckConfig("name-one"), nil)

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"in"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
}

func TestInfoCommandRunMissingArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{})
	require.Error(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Usage")
}

func TestInfoCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLIWithValue("none")
	client := cli.Client.(*client.MockClient)
	client.On("FetchCheck", "in").Return(types.FixtureCheckConfig("name-one"), nil)

	cmd := InfoCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"in"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Name")
	assert.Contains(out, "Interval")
	assert.Contains(out, "Command")
	assert.Contains(out, "Cron")
	assert.Contains(out, "Timeout")
	assert.Contains(out, "TTL")
	assert.Contains(out, "Subscriptions")
	assert.Contains(out, "Handlers")
	assert.Contains(out, "Runtime Assets")
	assert.Contains(out, "Hooks")
}

func TestInfoCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchCheck", "in").Return(&types.CheckConfig{}, errors.New("my-err"))

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"in"})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}
