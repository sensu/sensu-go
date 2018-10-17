package event

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
	client.On("ListEvents", mock.Anything).Return([]types.Event{
		*types.FixtureEvent("1", "something"),
		*types.FixtureEvent("2", "funny"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "something")
	assert.Contains(out, "funny")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithAllNamespaces(t *testing.T) {
	assert := assert.New(t)
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEvents", "*").Return([]types.Event{
		*types.FixtureEvent("1", "something"),
	}, nil)

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
	client.On("ListEvents", mock.Anything).Return([]types.Event{
		*types.FixtureEvent("1", "something"),
		*types.FixtureEvent("2", "funny"),
	}, nil)

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
	client.On("ListEvents", mock.Anything).Return([]types.Event{}, errors.New("fun-msg"))

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
