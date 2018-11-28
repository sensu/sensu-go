package check

import (
	"errors"
	"testing"

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

	cli := test.NewCLI()
	cmd := ListCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list", cmd.Use)
	assert.Regexp("checks", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListChecks", mock.Anything).Return([]types.CheckConfig{
		*types.FixtureCheckConfig("name-one"),
		*types.FixtureCheckConfig("name-two"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
	assert.Contains(out, "name-two")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithAll(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListChecks", "").Return([]types.CheckConfig{
		*types.FixtureCheckConfig("name-one"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	require.NoError(t, cmd.Flags().Set(flags.AllNamespaces, "t"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)
	assert.NotEmpty(out)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	check := types.FixtureCheckConfig("name-one")
	check.RuntimeAssets = []string{"asset-one"}

	client := cli.Client.(*client.MockClient)
	client.On("ListChecks", mock.Anything).Return([]types.CheckConfig{*check}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Name")      // heading
	assert.Contains(out, "Command")   // heading
	assert.Contains(out, "Assets")    // heading
	assert.Contains(out, "TTL")       // heading
	assert.Contains(out, "name-one")  // check name
	assert.Contains(out, "asset-one") // asset name
	assert.Contains(out, "60")        // interval
	assert.Contains(out, "Hooks")     // heading
}

// Test to ensure check command list output does not escape alphanumeric chars
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListChecks", mock.Anything).Return([]types.CheckConfig{}, errors.New("my-err"))

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
	checkConfig := types.FixtureCheckConfig("name-one")
	checkConfig.Command = "echo foo && exit 1"
	client.On("ListChecks", "").Return([]types.CheckConfig{
		*checkConfig,
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	require.NoError(t, cmd.Flags().Set(flags.AllNamespaces, "t"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)
	assert.NotEmpty(out)
	assert.Contains(out, "echo foo && exit 1")
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
