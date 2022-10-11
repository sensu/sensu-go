package subcommands

import (
	"testing"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSetCheckHooksCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetCheckHooksCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set-hooks", cmd.Use)
	assert.Regexp("of a check", cmd.Short)
}

func TestSetCheckHooksCommandRunEClosureSucess(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("AddCheckHook", mock.Anything, mock.Anything).Return(nil)
	client.On("FetchCheck", "name").Return(corev2.FixtureCheckConfig("name"), nil)

	cmd := SetCheckHooksCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", "non-zero"))
	require.NoError(t, cmd.Flags().Set("hooks", "hook"))

	out, err := test.RunCmd(cmd, []string{"name"})
	require.NoError(t, err)

	assert.Contains(out, "Added")
}

func TestSetCheckHooksCommandRunEInvalid(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := SetCheckHooksCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestSetCheckHooksCommandRunEClosureServerErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("AddCheckHook", mock.Anything, mock.Anything).Return(nil)

	cmd := SetCheckHooksCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestSetCheckHooksCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := SetCheckHooksCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Error(err)
}
