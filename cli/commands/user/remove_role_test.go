package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveRoleCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveRoleCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("remove-role", cmd.Use)
	assert.Regexp("remove role", cmd.Short)
}

func TestRemoveRoleCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveRoleCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestRemoveRoleCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveRoleFromUser", "foo", "bar").Return(nil)

	cmd := RemoveRoleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "bar"})

	assert.Regexp("Removed", out)
	assert.Nil(err)
}

func TestRemoveRoleCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveRoleFromUser", "bar", "foo").Return(errors.New("oh noes"))

	cmd := RemoveRoleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"bar", "foo"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("oh noes", err.Error())
}
