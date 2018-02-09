package silenced

import (
	"errors"
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("create", cmd.Use)
	assert.Regexp("silenced", cmd.Short)
}

func TestCreateCommandRunEClosureWithoutFlags(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateSilenced", mock.Anything).Return(fmt.Errorf("error"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("expire", "aaaaaa"))
	out, err := test.RunCmd(cmd, []string{"foo"})

	// Print help usage
	require.Error(t, err)
	assert.NotEmpty(out)
}

func TestCreateCommandRunEClosureWithAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateSilenced", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("reason", "just because"))
	require.NoError(t, cmd.Flags().Set("expire", "5"))
	require.NoError(t, cmd.Flags().Set("expire-on-resolve", "false"))
	require.NoError(t, cmd.Flags().Set("subscription", "weeklyworldnews"))
	require.NoError(t, cmd.Flags().Set("begin", "Jan 02 2006 3:04PM MST"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)
	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithDeps(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateSilenced", mock.AnythingOfType("*types.Silenced")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("reason", "just because"))
	require.NoError(t, cmd.Flags().Set("expire", "5"))
	require.NoError(t, cmd.Flags().Set("expire-on-resolve", "false"))
	require.NoError(t, cmd.Flags().Set("subscription", "weeklyworldnews"))

	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)
	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateSilenced", mock.AnythingOfType("*types.Silenced")).Return(errors.New("whoops"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("reason", "just because"))
	require.NoError(t, cmd.Flags().Set("expire", "5"))
	require.NoError(t, cmd.Flags().Set("expire-on-resolve", "false"))
	require.NoError(t, cmd.Flags().Set("subscription", "weeklyworldnews"))

	out, err := test.RunCmd(cmd, []string{})
	require.Error(t, err)
	assert.Equal("whoops", err.Error())
	assert.Empty(out)
}

func TestCreateCommandRunEClosureWithMissingRequiredFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateSilenced", mock.AnythingOfType("*types.Silenced")).Return(errors.New("error"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("expire", "5"))
	out, err := test.RunCmd(cmd, []string{})
	require.Error(t, err)
	assert.Empty(out)
}
