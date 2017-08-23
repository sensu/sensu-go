package user

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestDeleteCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("disable", cmd.Use)
	assert.Regexp("disable user", cmd.Short)
}

func TestDeleteCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	cmd.Flags().Set("timeout", "15")
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Nil(err)
}

func TestDeleteCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DisableUser", "foo").Return(nil)

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"foo"})

	assert.Regexp("Disabled", out)
	assert.Nil(err)
}

func TestDeleteCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DisableUser", "bar").Return(errors.New("oh noes"))

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"bar"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}

func TestDeleteCommandRunEFailConfirm(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"username"})

	assert.Contains(out, "Canceled")
	assert.NoError(err)
}
