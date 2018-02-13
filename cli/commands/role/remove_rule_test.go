package role

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestRemoveRuleCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveRuleCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("remove-rule", cmd.Use)
	assert.Regexp("remove rule given name", cmd.Short)
}

func TestRemoveRuleCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveRuleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"sdfasdf"})

	// Print help usage
	assert.NotEmpty(out)
	assert.Error(err)
}

func TestRemoveRuleCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveRule", "foo", "bar").Return(nil)

	cmd := RemoveRuleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "bar"})

	assert.Regexp("Removed", out)
	assert.Nil(err)
}

func TestRemoveRuleCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveRule", "bar", "foo").Return(errors.New("oh noes"))

	cmd := RemoveRuleCommand(cli)
	out, err := test.RunCmd(cmd, []string{"bar", "foo"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}
