package hook

import (
	"errors"
	"net/http"
	"testing"

	corev2 "github.com/sensu/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("hooks", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.HookConfig{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.HookConfig)
			*resources = []corev2.HookConfig{
				*corev2.FixtureHookConfig("name-one"),
				*corev2.FixtureHookConfig("name-two"),
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
	assert.Contains(out, "name-two")
	assert.Nil(err)
	assert.NotContains(out, "==")
}

func TestListCommandRunEClosureWithAll(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.HookConfig{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.HookConfig)
			*resources = []corev2.HookConfig{
				*corev2.FixtureHookConfig("name-one"),
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	require.NoError(t, cmd.Flags().Set(flags.AllNamespaces, "t"))
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Nil(err)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	hook := corev2.FixtureHookConfig("name-one")

	client := cli.Client.(*client.MockClient)
	resources := []corev2.HookConfig{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.HookConfig)
			*resources = []corev2.HookConfig{
				*hook,
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Name")     // heading
	assert.Contains(out, "Command")  // heading
	assert.Contains(out, "name-one") // hook name
	assert.Contains(out, "10")       // timeout
	assert.Nil(err)
}

// Test to ensure hook command list output does not escape alphanumeric chars
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.HookConfig{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(errors.New("my-err"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
	assert.Empty(out)
}

func TestListCommandRunEClosureWithAlphaNumericChars(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	hookConfig := corev2.FixtureHookConfig("name-one")
	hookConfig.Command = "echo foo && exit 1"
	resources := []corev2.HookConfig{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.HookConfig)
			*resources = []corev2.HookConfig{
				*hookConfig,
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	require.NoError(t, cmd.Flags().Set(flags.AllNamespaces, "t"))
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Contains(out, "echo foo && exit 1")
	assert.Nil(err)
}

func TestListFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := ListCommand(cli)

	flag := cmd.Flag("all-namespaces")
	assert.NotNil(flag)

	flag = cmd.Flag("format")
	assert.NotNil(flag)
}

func TestListCommandRunEClosureWithHeader(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	var header http.Header
	resources := []corev2.HookConfig{}
	client.On("List", mock.Anything, &resources, mock.Anything, &header).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.HookConfig)
			*resources = []corev2.HookConfig{}
			header := args[3].(*http.Header)
			*header = make(http.Header)
			header.Add(helpers.HeaderWarning, "E_TOO_MANY_ENTITIES")
		},
	)

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
	assert.Contains(out, "E_TOO_MANY_ENTITIES")
	assert.Contains(out, "==")
}
