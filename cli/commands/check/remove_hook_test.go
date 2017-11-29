package check

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestRemoveCheckHookCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveCheckHookCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("remove-hook", cmd.Use)
	assert.Regexp("remove hook from check", cmd.Short)
}

func TestRemoveCheckHookCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"sdfasdf"})

	assert.Empty(out)
	assert.Error(err)
}

func TestRemoveCheckHookCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveCheckHook", "foo", "bar", "baz").Return(nil)

	cmd := RemoveCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"foo", "bar", "baz"})

	assert.Regexp("Removed", out)
	assert.Nil(err)
}

func TestRemoveCheckHookCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveCheckHook", "bar", "foo", "baz").Return(errors.New("oh noes"))

	cmd := RemoveCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"bar", "foo", "baz"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}
