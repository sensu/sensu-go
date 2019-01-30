package role

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
	assert.Regexp("role", cmd.Short)
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

func TestCreateCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	cli.Client.(*client.MockClient).
		On("CreateRole", mock.Anything).
		Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("verb", "list"))
	require.NoError(t, cmd.Flags().Set("resource", "events"))
	out, err := test.RunCmd(cmd, []string{"foo"})

	assert.Regexp("Created", out)
	assert.NoError(err)
}
