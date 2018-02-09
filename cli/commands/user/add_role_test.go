package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRoleCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := AddRoleCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("add-role", cmd.Use)
	assert.Regexp("add role", cmd.Short)
}

func TestAddRoleCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := AddRoleCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.NoError(err)         // no error if help usage prints
}

func TestAddRoleCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("AddRoleToUser", "foo", "bar").Return(nil)

	cmd := AddRoleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "bar"})

	assert.Regexp("Added", out)
	assert.Nil(err)
}

func TestAddRoleCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("AddRoleToUser", "bar", "foo").Return(errors.New("oh noes"))

	cmd := AddRoleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"bar", "foo"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("oh noes", err.Error())
}
