package role

import (
	"fmt"
	"testing"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("create", cmd.Use)
	assert.Regexp("roles", cmd.Short)
}

func TestListCommandRunEClosureSucess(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("CreateRole", mock.AnythingOfType("*types.Role")).Return(nil)

	cmd := CreateCommand(cli)
	out, err := test.RunCmd(cmd, []string{"my-name"})

	assert.Contains(out, "Created")
	assert.NoError(err)
}

func TestListCommandRunEClosureServerErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("CreateRole", mock.AnythingOfType("*types.Role")).Return(fmt.Errorf(""))

	cmd := CreateCommand(cli)
	out, err := test.RunCmd(cmd, []string{"asdfasdfad"})

	assert.Empty(out)
	assert.Error(err)
}

func TestListCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := CreateCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	// Print help usage
	assert.NotEmpty(out)
	assert.Error(err)
}
