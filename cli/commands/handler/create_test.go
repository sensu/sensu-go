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
	client.On("CreateHandler", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", "set"))
	require.NoError(t, cmd.Flags().Set("timeout", "15"))
	require.NoError(t, cmd.Flags().Set("mutator", ""))
	require.NoError(t, cmd.Flags().Set("handlers", "slack,pagerduty"))
	require.NoError(t, cmd.Flags().Set("env-vars", "key1=val1,key2=val2"))
	out, err := test.RunCmd(cmd, []string{"test-handler"})

	assert.Regexp("Created", out)
	assert.Nil(err)
}

func TestCreateCommandRunEClosureWithAPIErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateHandler", mock.Anything).Return(errors.New("nope"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("type", "set"))
	require.NoError(t, cmd.Flags().Set("timeout", "15"))
	require.NoError(t, cmd.Flags().Set("mutator", ""))
	require.NoError(t, cmd.Flags().Set("handlers", "slack,pagerduty"))
	out, err := test.RunCmd(cmd, []string{"nope-jpeg"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("nope", err.Error())
}
