package role

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()
	cmd := ListCommand(cli)
	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("roles", cmd.Short)
}
func TestListCommandRunEClosureJSONFormat(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListRoles").Return([]types.Role{
		*types.FixtureRole("one", "*"),
		*types.FixtureRole("two", "*"),
	}, nil)
	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Nil(err)
}
func TestListCommandRunEClosureTabularFormat(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("")
	client := cli.Client.(*client.MockClient)
	client.On("ListRoles").Return([]types.Role{
		*types.FixtureRole("one", "*"),
		*types.FixtureRole("two", "*"),
	}, nil)
	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Contains(out, "Name")
	assert.Contains(out, "one")
	assert.Contains(out, "two")
	assert.Nil(err)
}
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListRoles").Return([]types.Role{}, errors.New("fire"))
	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})
	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("fire", err.Error())
}
