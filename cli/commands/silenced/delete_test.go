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

func TestDeleteCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteSilenced", mock.AnythingOfType("string")).Return(nil)
	cmd := DeleteCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("delete", cmd.Use)
	assert.Regexp("silenced", cmd.Short)
}

func TestDeleteCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteSilenced", mock.AnythingOfType("string")).Return(nil)
	cmd := DeleteCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	require.NoError(t, cmd.Flags().Set("subscription", "nytimes"))
	out, err := test.RunCmd(cmd, []string{"a", "b"})

	require.Error(t, err)
	assert.Regexp("Usage", out) // usage should print out
}

func TestDeleteCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteSilenced", mock.AnythingOfType("string")).Return(nil)

	cmd := DeleteCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{"foo:bar"})

	require.NoError(t, err)
	assert.Regexp("OK", out)
}

func TestDeleteCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteSilenced", mock.AnythingOfType("string")).Return(errors.New("oh noes"))

	cmd := DeleteCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{"foo:bar"})

	require.Error(t, err)
	assert.Empty(out)
	assert.Equal("oh noes", err.Error())
}

func TestDeleteCommandRunEFailConfirm(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	out, err := test.RunCmd(cmd, []string{"test-handler"})

	require.NoError(t, err)
	assert.Contains(out, "Canceled")
}
