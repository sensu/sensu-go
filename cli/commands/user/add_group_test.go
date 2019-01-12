package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddGroupCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := AddGroupCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("add-group", cmd.Use)
	assert.Regexp("add group", cmd.Short)
}

func TestAddGroupCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := AddGroupCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestAddGroupCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("AddGroupToUser", "user", "group").Return(nil)

	cmd := AddGroupCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user", "group"})

	assert.Regexp("Added", out)
	assert.Nil(err)
}

func TestAddGroupCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("AddGroupToUser", "user", "group").Return(errors.New("failure"))

	cmd := AddGroupCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user", "group"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("failure", err.Error())
}
