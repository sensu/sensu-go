package event

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/cli"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
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
	client.On("ListEvents").Return([]types.Event{
		*types.FixtureEvent("1", "something"),
		*types.FixtureEvent("2", "funny"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "json")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "something")
	assert.Contains(out, "funny")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEvents").Return([]types.Event{
		*types.FixtureEvent("1", "something"),
		*types.FixtureEvent("2", "funny"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "tabular")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Source")    // Heading
	assert.Contains(out, "Check")     // Heading
	assert.Contains(out, "Result")    // Heading
	assert.Contains(out, "Timestamp") // Heading
	assert.Contains(out, "something")
	assert.Contains(out, "funny")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)
	cli := newConfiguredCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEvents").Return([]types.Event{}, errors.New("fun-msg"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("fun-msg", err.Error())
}

func newConfiguredCLI() *cli.SensuCli {
	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("json")
	return cli
}
