package apikey

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/cli"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRevokeCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RevokeCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("revoke", cmd.Use)
	assert.Regexp("api-key", cmd.Short)
}

func TestRevokeCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RevokeCommand(cli)
	require.Error(t, cmd.Flags().Set("username", "user1"))
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out)
	assert.Error(err)
}

func TestRevokeCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("Delete", mock.Anything, mock.Anything).Return(nil)

	cmd := RevokeCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{"my-api-key"})

	assert.Regexp("Deleted", out)
	assert.Nil(err)
}

func TestDeleteCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("Delete", mock.Anything, mock.Anything).Return(errors.New("err"))

	cmd := RevokeCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{"my-api-key"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("err", err.Error())
}

func TestDeleteCommandRunEFailConfirm(t *testing.T) {
	assert := assert.New(t)

	test.WithMockCLI(t, func(cli *cli.SensuCli) {
		cmd := RevokeCommand(cli)
		output, err := test.RunCmdWithOutFile(cmd, []string{"my-api-key"}, cli.OutFile)

		require.NoError(t, err)
		assert.Contains(output, "Canceled")
	})
}
