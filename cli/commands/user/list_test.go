package user

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/cli"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("user", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListUsers").Return([]types.User{
		*types.FixtureUser("one"),
		*types.FixtureUser("two"),
	}, nil)

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListUsers").Return([]types.User{}, errors.New("fire"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := newCLI()

	client := cli.Client.(*client.MockClient)
	client.On("ListUsers").Return([]types.User{
		*types.FixtureUser("one"),
		*types.FixtureUser("two"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Username")
	assert.Contains(out, "Roles")
	assert.Contains(out, "Enabled")
	assert.Contains(out, "one")
	assert.Contains(out, "two")
	assert.Contains(out, "true")
	assert.NoError(err)
}

func newCLI() *cli.SensuCli {
	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("json")

	return cli
}
