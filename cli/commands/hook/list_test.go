package hook

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/cli"
	client "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/sensu/sensu-go/cli/commands/flags"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("hooks", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListHooks", mock.Anything).Return([]types.HookConfig{
		*types.FixtureHookConfig("name-one"),
		*types.FixtureHookConfig("name-two"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "json")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
	assert.Contains(out, "name-two")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithAll(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListHooks", "*").Return([]types.HookConfig{
		*types.FixtureHookConfig("name-one"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set(flags.Format, "json")
	cmd.Flags().Set(flags.AllOrgs, "t")
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Nil(err)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := newCLI()

	hook := types.FixtureHookConfig("name-one")

	client := cli.Client.(*client.MockClient)
	client.On("ListHooks", mock.Anything).Return([]types.HookConfig{*hook}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "none")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Name")     // heading
	assert.Contains(out, "Command")  // heading
	assert.Contains(out, "name-one") // hook name
	assert.Contains(out, "60")       // timeout
	assert.Nil(err)
}

// Test to ensure hook command list output does not escape alphanumeric chars
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListHooks", mock.Anything).Return([]types.HookConfig{}, errors.New("my-err"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}

func TestListCommandRunEClosureWithAlphaNumericChars(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	hookConfig := types.FixtureHookConfig("name-one")
	hookConfig.Command = "echo foo && exit 1"
	client.On("ListHooks", "*").Return([]types.HookConfig{
		*hookConfig,
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set(flags.Format, "json")
	cmd.Flags().Set(flags.AllOrgs, "t")
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Contains(out, "echo foo && exit 1")
	assert.Nil(err)
}

func TestListFlags(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	cmd := ListCommand(cli)

	flag := cmd.Flag("all-organizations")
	assert.NotNil(flag)

	flag = cmd.Flag("format")
	assert.NotNil(flag)
}

func newCLI() *cli.SensuCli {
	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("json")

	return cli
}
