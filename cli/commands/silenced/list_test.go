package silenced

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

	cli := newCLI()
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("silenced entries", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("foo:bar"),
		*types.FixtureSilenced("bar:foo"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "json")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "foo:bar")
	assert.Contains(out, "bar:foo")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithAll(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything).Return([]types.Silenced{
		*types.FixtureSilenced("foo:bar"),
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

	silenced := types.FixtureSilenced("foo:bar")
	silenced.Reason = "justcause!"
	silenced.Creator = "eric"
	silenced.Check = "bar"
	silenced.Subscription = "foo"
	silenced.Organization = "defaultorg"
	silenced.Environment = "defaultenv"

	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything).Return([]types.Silenced{*silenced}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set("format", "none")
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "ID")              // heading
	assert.Contains(out, "Expire")          // heading
	assert.Contains(out, "ExpireOnResolve") // heading
	assert.Contains(out, "Creator")         // heading
	assert.Contains(out, "Check")           // heading
	assert.Contains(out, "Reason")          // heading
	assert.Contains(out, "Subscription")    // heading
	assert.Contains(out, "Organization")    // heading
	assert.Contains(out, "Environment")     // heading
	assert.Contains(out, "justcause!")
	assert.Contains(out, "foo:bar")
	assert.Contains(out, "eric")
	assert.Contains(out, "defaultorg")
	assert.Contains(out, "defaultenv")
	assert.Contains(out, "false")
	assert.Contains(out, "0s")
}

// Test to ensure silenced command list output does not escape alphanumeric chars
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything).Return([]types.Silenced{}, errors.New("my-err"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
	assert.Empty(out)
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
