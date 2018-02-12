package subcommands

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRemoveCheckHookCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveCheckHookCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("remove-hook", cmd.Use)
	assert.Regexp("from a check", cmd.Short)
}

func TestRemoveCheckHookCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RemoveCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"sdfasdf"})

	assert.NotEmpty(out)
	assert.Error(err)
}

func TestRemoveCheckHookCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveCheckHook", mock.AnythingOfType("*types.CheckConfig"), "non-zero", "hook1").Return(nil)
	client.On("FetchCheck", "name").Return(types.FixtureCheckConfig("name"), nil)

	cmd := RemoveCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name", "non-zero", "hook1"})

	assert.Regexp("Removed", out)
	assert.Nil(err)
}

func TestRemoveCheckHookCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RemoveCheckHook", mock.AnythingOfType("*types.CheckConfig"), "non-zero", "hook1").Return(errors.New("oh noes"))
	client.On("FetchCheck", "name").Return(types.FixtureCheckConfig("name"), nil)

	cmd := RemoveCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name", "non-zero", "hook1"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}
