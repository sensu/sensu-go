package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetGroupsCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetGroupsCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set-groups", cmd.Use)
	assert.Regexp("set .* groups", cmd.Short)
}

func TestSetGroupsCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetGroupsCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestSetGroupsCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("SetGroupsForUser", "user", []string{"group1", "group2", "group3"}).Return(nil)

	cmd := SetGroupsCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user", "group1,group2,group3"})

	assert.Regexp("Set", out)
	assert.Nil(err)
}

func TestSetGroupsCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("SetGroupsForUser", "user", []string{"group1"}).Return(errors.New("failure"))

	cmd := SetGroupsCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user", "group1"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("failure", err.Error())
}
