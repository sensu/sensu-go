package silenced

import (
	"errors"
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
	client.On("CreateSilenced", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)

	out, err := test.RunCmd(cmd, []string{"foo"})
	require.Error(t, err)
	assert.Empty(out)
}

func TestCreateCommandRunEClosureWithAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateSilenced", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	cmd.Flags().Set("expire", "5")
	cmd.Flags().Set("expire-on-resolve", "false")
	cmd.Flags().Set("subscription", "weeklyworldnews")

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
	cmd.Flags().Set("expire", "5")
	cmd.Flags().Set("expire-on-resolve", "false")
	cmd.Flags().Set("subscription", "weeklyworldnews")

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
	cmd.Flags().Set("expire", "5")
	cmd.Flags().Set("expire-on-resolve", "false")
	cmd.Flags().Set("subscription", "weeklyworldnews")

	out, err := test.RunCmd(cmd, []string{})
	require.Error(t, err)
	assert.Equal("whoops", err.Error())
	assert.Empty(out)
}
