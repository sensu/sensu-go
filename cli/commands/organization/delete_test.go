package organization

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
	cmd.Flags().Set("skip-confirm", "t")

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("delete", cmd.Use)
	assert.Regexp("organization", cmd.Short)
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
	client.On("DeleteOrganization", "foo").Return(nil)

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"foo"})

	assert.Regexp("Deleted", out)
	assert.Nil(err)
}

func TestDeleteCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteOrganization", "bar").Return(errors.New("oh noes"))

	cmd := DeleteCommand(cli)
	cmd.Flags().Set("skip-confirm", "t")
	out, err := test.RunCmd(cmd, []string{"bar"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}
