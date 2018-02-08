package check

import (
	"errors"
	"testing"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ExecuteCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("execute", cmd.Use)
	assert.Regexp("request", cmd.Short)
}

func TestExecuteCommandRunEClosureSuccess(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("ExecuteCheck", mock.AnythingOfType("*types.AdhocRequest")).Return(nil)

	cmd := ExecuteCommand(cli)
	require.NoError(t, cmd.Flags().Set("reason", "foo"))

	out, err := test.RunCmd(cmd, []string{"name"})
	require.NoError(t, err)

	assert.Contains(out, "Issued")
}

func TestExecuteCommandRunEInvalid(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := ExecuteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestExecuteCommandRunEClosureServerErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("ExecuteCheck", mock.AnythingOfType("*types.AdhocRequest")).Return(errors.New("whoops"))

	cmd := ExecuteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"name"})

	assert.Empty(out)
	assert.Error(err)
}

func TestExecuteCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := ExecuteCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}
