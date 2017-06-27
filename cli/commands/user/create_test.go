package user

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
	assert.Regexp("users", cmd.Short)
}

func TestListCommandRunEClosureWithArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("CreateUser", mock.AnythingOfType("*types.User")).Return(nil)

	cmd := CreateCommand(cli)
	cmd.Flags().Set("username", "bob")
	cmd.Flags().Set("password", "b0b")

	out, err := test.RunCmd(cmd, []string{})

	assert.Contains(out, "Created")
	assert.NoError(err)
}

func TestListCommandRunEClosureServerErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("CreateUser", mock.AnythingOfType("*types.User")).Return(fmt.Errorf(""))

	cmd := CreateCommand(cli)
	cmd.Flags().Set("username", "bob")
	cmd.Flags().Set("password", "b0b")

	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}

func TestListCommandRunEClosureWithRoles(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	client := cli.Client.(*clientmock.MockClient)
	client.On("CreateUser", mock.AnythingOfType("*types.User")).Return(nil)

	cmd := CreateCommand(cli)
	cmd.Flags().Set("username", "bob")
	cmd.Flags().Set("password", "b0b")
	cmd.Flags().Set("roles", "    default,   meowmix    ")

	out, err := test.RunCmd(cmd, []string{})

	assert.Contains(out, "Created")
	assert.NoError(err)
}

func TestListCommandRunEClosureMissingArgs(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	cmd := CreateCommand(cli)
	cmd.Flags().Set("username", "")
	cmd.Flags().Set("password", "b0b")

	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}
