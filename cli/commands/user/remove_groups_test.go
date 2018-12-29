package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveAllGroupsCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveAllGroupsCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("remove-groups", cmd.Use)
	assert.Regexp("remove all the groups", cmd.Short)
}

func TestRemoveAllGroupsCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveAllGroupsCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestRemoveAllGroupsCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveAllGroupsFromUser", "user").Return(nil)

	cmd := RemoveAllGroupsCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestRemoveAllGroupsCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveAllGroupsFromUser", "user").Return(errors.New("failure"))

	cmd := RemoveAllGroupsCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("failure", err.Error())
}
