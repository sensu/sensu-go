package asset

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("create", cmd.Use)
	assert.Regexp("assets", cmd.Short)
}

func TestCreateCommandRunEClosureWithoutFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)
	out, err := test.RunCmd(cmd, []string{"my-asset"})

	assert.Empty(out)
	assert.NotNil(err)
}

func TestCreateCommandRunEClosureWithAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateAsset", mock.AnythingOfType("*types.Asset")).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("url", "http://lol"))
	require.NoError(t, cmd.Flags().Set("sha512", "12345qwerty"))
	out, err := test.RunCmd(cmd, []string{"ruby22"})
	require.NoError(t, err)

	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateAsset", mock.AnythingOfType("*types.Asset")).Return(errors.New("whoops"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("sha512", "12345qwerty"))
	require.NoError(t, cmd.Flags().Set("url", "http://lol"))
	out, err := test.RunCmd(cmd, []string{"ruby22"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("whoops", err.Error())
}

func TestCreateExectorBadURLGiven(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	require.NoError(t, cmd.Flags().Set("url", "my-bad-bad-url-boy"))
	exec := &CreateExecutor{Client: cli.Client}

	err := exec.Run(cmd, []string{"ruby22"})
	assert.Error(err)
}

func TestConfigureAsset(t *testing.T) {
	assert := assert.New(t)

	flags := &pflag.FlagSet{}
	flags.StringSlice("metadata", []string{}, "")
	flags.StringSlice("filter", []string{}, "")
	flags.String("sha512", "12345qwerty", "")
	flags.String("url", "http://lol", "")

	// Too many args
	cfg := ConfigureAsset{Flags: flags, Args: []string{"one", "too many"}, Org: "default"}
	asset, errs := cfg.Configure()
	assert.NotEmpty(errs)
	assert.Empty(asset.Name)

	// Empty org
	cfg = ConfigureAsset{Flags: flags, Args: []string{"ruby22"}, Org: ""}
	asset, errs = cfg.Configure()
	assert.NotEmpty(errs)
	assert.Empty(asset.Organization)

	// Valid Metadata
	require.NoError(t, flags.Set("metadata", "One: Two"))
	require.NoError(t, flags.Set("metadata", "  Three : Four "))
	cfg = ConfigureAsset{Flags: flags, Args: []string{"ruby22"}, Org: "default"}
	asset, errs = cfg.Configure()
	assert.Empty(errs)
	assert.Equal("ruby22", asset.Name)
	assert.NotEmpty(asset.Metadata)
	assert.Equal("Two", asset.Metadata["One"])

	// Bad Metadata
	require.NoError(t, flags.Set("metadata", "Five- Six"))
	_, errs = cfg.Configure()
	assert.NotEmpty(errs)
}
