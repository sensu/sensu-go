package event

import (
	"errors"
	"net/http"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
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

	cli := newConfiguredCLI()
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("events", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.Event{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.Event)
			*resources = []corev2.Event{
				*corev2.FixtureEvent("1", "something"),
				*corev2.FixtureEvent("2", "funny"),
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "something")
	assert.Contains(out, "funny")
	assert.Nil(err)
	assert.NotContains(out, "==")
}

func TestListCommandRunEClosureWithAllNamespaces(t *testing.T) {
	assert := assert.New(t)
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.Event{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.Event)
			*resources = []corev2.Event{
				*corev2.FixtureEvent("1", "something"),
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
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.Event{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.Event)
			*resources = []corev2.Event{
				*corev2.FixtureEvent("1", "something"),
				*corev2.FixtureEvent("2", "funny"),
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "none"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Entity")    // Heading
	assert.Contains(out, "Check")     // Heading
	assert.Contains(out, "Output")    // Heading
	assert.Contains(out, "Timestamp") // Heading
	assert.Contains(out, "something")
	assert.Contains(out, "funny")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.Event{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(errors.New("fun-msg"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("fun-msg", err.Error())
}

func TestListFlags(t *testing.T) {
	assert := assert.New(t)

	cli := newConfiguredCLI()
	cmd := ListCommand(cli)

	flag := cmd.Flag("all-namespaces")
	assert.NotNil(flag)

	flag = cmd.Flag("format")
	assert.NotNil(flag)
}

func newConfiguredCLI() *cli.SensuCli {
	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("json")
	return cli
}

func TestListCommandRunEClosureWithHeader(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	var header http.Header
	resources := []corev2.Event{}
	client.On("List", mock.Anything, &resources, mock.Anything, &header).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.Event)
			*resources = []corev2.Event{}
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
