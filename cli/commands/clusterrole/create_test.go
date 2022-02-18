package clusterrole

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestCreateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("create", cmd.Use)
	assert.Regexp("cluster role", cmd.Short)
}

func TestCreateCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := CreateCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	// Print help usage
	assert.NotEmpty(out)
	assert.Error(err)
}

func TestCreateCommandDeprecatedRunEClosureWithDeps(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateClusterRole", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("verb", "list"))
	require.NoError(t, cmd.Flags().Set("resource", "events"))
	out, err := test.RunCmd(cmd, []string{"my-role"})

	assert.NoError(err)
	assert.Regexp("Created", out)
}

func TestCreateCommandRunEClosureWithDeps(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateClusterRole", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("verbs", "list"))
	require.NoError(t, cmd.Flags().Set("resources", "events"))
	out, err := test.RunCmd(cmd, []string{"my-role"})

	assert.NoError(err)
	assert.Regexp("Created", out)
}
