package handler

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
	assert.Regexp("handlers", cmd.Short)
}

func TestCreateCommandRunEClosureWithoutAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", ""))
	out, err := test.RunCmd(cmd, []string{"my-handler"})
	require.Error(t, err)
	assert.Regexp("Usage", out) // usage should print out
}

func TestCreateCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateHandler", mock.AnythingOfType("*types.Handler")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", "pipe"))
	require.NoError(t, cmd.Flags().Set("timeout", "15"))
	require.NoError(t, cmd.Flags().Set("mutator", ""))
	require.NoError(t, cmd.Flags().Set("handlers", "slack,pagerduty"))
	out, err := test.RunCmd(cmd, []string{"test-handler"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestCreateCommandRunEClosureWithAPIErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateHandler", mock.AnythingOfType("*types.Handler")).Return(errors.New("nope"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", "pipe"))
	require.NoError(t, cmd.Flags().Set("timeout", "15"))
	require.NoError(t, cmd.Flags().Set("mutator", ""))
	require.NoError(t, cmd.Flags().Set("handlers", "slack,pagerduty"))
	out, err := test.RunCmd(cmd, []string{"nope-jpeg"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("nope", err.Error())
}
