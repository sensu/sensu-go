package check

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
	assert.Regexp("checks", cmd.Short)
}

func TestCreateCommandRunEClosureWithoutFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("interval", "sdfsa"))
	out, err := test.RunCmd(cmd, []string{"echo 'heyhey'"})

	assert.Empty(out)
	assert.NotNil(err)
}

func TestCreateCommandRunEClosureWithAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateCheck", mock.AnythingOfType("*types.CheckConfig")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("command", "echo 'heyhey'"))
	require.NoError(t, cmd.Flags().Set("subscriptions", "system"))
	require.NoError(t, cmd.Flags().Set("interval", "10"))
	out, err := test.RunCmd(cmd, []string{"can-holla"})
	require.NoError(t, err)

	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithDeps(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateCheck", mock.AnythingOfType("*types.CheckConfig")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("command", "echo 'heyhey'"))
	require.NoError(t, cmd.Flags().Set("subscriptions", "system"))
	require.NoError(t, cmd.Flags().Set("interval", "10"))
	require.NoError(t, cmd.Flags().Set("runtime-assets", "ruby22"))
	out, err := test.RunCmd(cmd, []string{"can-holla"})
	require.NoError(t, err)

	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithDepsSTDIN(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateCheck", mock.AnythingOfType("*types.CheckConfig")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("command", "echo 'heyhey'"))
	require.NoError(t, cmd.Flags().Set("subscriptions", "system"))
	require.NoError(t, cmd.Flags().Set("interval", "10"))
	require.NoError(t, cmd.Flags().Set("runtime-assets", "ruby22"))
	require.NoError(t, cmd.Flags().Set("stdin", "true"))
	out, err := test.RunCmd(cmd, []string{"can-holla"})
	require.NoError(t, err)

	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateCheck", mock.AnythingOfType("*types.CheckConfig")).Return(errors.New("whoops"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("command", "echo 'heyhey'"))
	require.NoError(t, cmd.Flags().Set("subscriptions", "system"))
	require.NoError(t, cmd.Flags().Set("interval", "10"))
	out, err := test.RunCmd(cmd, []string{"can-holla"})

	assert.Empty(out)
	assert.Error(err)
	assert.Equal("whoops", err.Error())
}

func TestCreateCommandRunEClosureWithMissingRequiredFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateCheck", mock.AnythingOfType("*types.CheckConfig")).Return(errors.New("error"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("interval", "5"))
	out, err := test.RunCmd(cmd, []string{"checky"})
	require.Error(t, err)
	assert.Empty(out)
}
