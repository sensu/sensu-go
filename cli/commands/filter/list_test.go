package filter

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
	assert.Regexp("filters", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListFilters", mock.Anything).Return([]types.EventFilter{
		*types.FixtureEventFilter("name-one"),
		*types.FixtureEventFilter("name-two"),
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
	client.On("ListFilters", "*").Return([]types.EventFilter{
		*types.FixtureEventFilter("name-one"),
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

	filter := types.FixtureEventFilter("name-one")

	client := cli.Client.(*client.MockClient)
	client.On("ListFilters", mock.Anything).Return([]types.EventFilter{*filter}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "none")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Name")       // heading
	assert.Contains(out, "Action")     // heading
	assert.Contains(out, "Statements") // heading
	assert.Nil(err)
}

// Test to ensure check command list output does not escape alphanumeric chars
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListFilters", mock.Anything).Return([]types.EventFilter{}, errors.New("my-err"))

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
	filter := types.FixtureEventFilter("name-one")
	filter.Statements = append(filter.Statements, "10 > 0")
	client.On("ListFilters", "*").Return([]types.EventFilter{*filter}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set(flags.Format, "json")
	cmd.Flags().Set(flags.AllOrgs, "t")
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Contains(out, "10 > 0")
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
