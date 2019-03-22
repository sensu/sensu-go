package clusterrolebinding

import (
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
	cmd := DeleteCommand(cli)
	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("delete", cmd.Use)
	assert.Regexp("cluster role binding", cmd.Short)
}
func TestDeleteCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	cmd := DeleteCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{})
	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}
func TestDeleteCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeleteClusterRoleBinding", mock.Anything, mock.Anything).Return(nil)
	cmd := DeleteCommand(cli)
	require.NoError(t, cmd.Flags().Set("skip-confirm", "t"))
	out, err := test.RunCmd(cmd, []string{"foo"})
	assert.Regexp("Deleted", out)
	assert.Nil(err)
}
