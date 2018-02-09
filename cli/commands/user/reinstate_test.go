package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReinstateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ReinstateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("reinstate", cmd.Use)
	assert.Regexp("reinstate disabled user", cmd.Short)
}

func TestReinstateCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ReinstateCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestReinstateCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ReinstateUser", "foo").Return(nil)

	cmd := ReinstateCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo"})

	assert.Regexp("Reinstated", out)
	assert.Nil(err)
}

func TestReinstateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ReinstateUser", "bar").Return(errors.New("oh noes"))

	cmd := ReinstateCommand(cli)
	out, err := test.RunCmd(cmd, []string{"bar"})

	assert.Empty(out)
	require.Error(t, err)
	assert.Equal("oh noes", err.Error())
}
