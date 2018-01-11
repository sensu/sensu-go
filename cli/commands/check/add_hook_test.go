package check

import (
	"testing"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAddCheckHookCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := AddCheckHookCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("add-hook", cmd.Use)
	assert.Regexp("to check", cmd.Short)
}

func TestAddCheckHookCommandRunEClosureSucess(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("AddCheckHook", mock.AnythingOfType("*types.CheckConfig"), mock.AnythingOfType("*types.HookList")).Return(nil)
	client.On("FetchCheck", "name").Return(types.FixtureCheckConfig("name"), nil)

	cmd := AddCheckHookCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", "non-zero"))

	out, err := test.RunCmd(cmd, []string{"name"})
	require.NoError(t, err)

	assert.Contains(out, "Added")
}

func TestAddCheckHookCommandRunEInvalid(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := AddCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestAddCheckHookCommandRunEClosureServerErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("AddCheckHook", mock.AnythingOfType("*types.CheckConfig"), mock.AnythingOfType("*types.checkHook")).Return(nil)

	cmd := AddCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestAddCheckHookCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := AddCheckHookCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}
