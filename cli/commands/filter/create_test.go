package filter

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
	assert.Regexp("filters", cmd.Short)
}

func TestCreateCommandRunEClosureWithoutFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("action", "allow"))
	out, err := test.RunCmd(cmd, []string{"echo 'heyhey'"})

	assert.Empty(out)
	assert.NotNil(err)
}

func TestCreateCommandRunEClosureWithAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateFilter", mock.AnythingOfType("*types.EventFilter")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("action", "allow"))
	require.NoError(t, cmd.Flags().Set("expressions", "10 > 0"))
	out, err := test.RunCmd(cmd, []string{"can-holla"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestCreateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateFilter", mock.AnythingOfType("*types.EventFilter")).Return(errors.New("whoops"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("action", "allow"))
	require.NoError(t, cmd.Flags().Set("expressions", "10 > 0"))
	out, err := test.RunCmd(cmd, []string{"can-holla"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("whoops", err.Error())
}
