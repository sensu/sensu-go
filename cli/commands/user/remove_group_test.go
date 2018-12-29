package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveGroupCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveGroupCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("remove-group", cmd.Use)
	assert.Regexp("remove group", cmd.Short)
}

func TestRemoveGroupCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveGroupCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestRemoveGroupCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveGroupFromUser", "user", "group").Return(nil)

	cmd := RemoveGroupCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user", "group"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestRemoveGroupCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveGroupFromUser", "user", "group").Return(errors.New("failure"))

	cmd := RemoveGroupCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user", "group"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("failure", err.Error())
}
