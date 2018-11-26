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
	client.On("CreateAsset", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("url", "http://lol"))
	require.NoError(t, cmd.Flags().Set("sha512", "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17"))
	out, err := test.RunCmd(cmd, []string{"ruby22"})
	require.NoError(t, err)

	assert.Regexp("OK", out)
}

func TestCreateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateAsset", mock.Anything).Return(errors.New("whoops"))

	cmd := CreateCommand(cli)
	require.NoError(t, cmd.Flags().Set("sha512", "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17"))
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
	flags.StringSlice("filter", []string{}, "")
	flags.String("sha512", "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17", "")
	flags.String("url", "http://lol", "")

	// Too many args
	cfg := ConfigureAsset{Flags: flags, Args: []string{"one", "too many"}, Namespace: "default"}
	asset, errs := cfg.Configure()
	assert.NotEmpty(errs)
	assert.Empty(asset.Name)

	// Empty namespace
	cfg = ConfigureAsset{Flags: flags, Args: []string{"ruby22"}, Namespace: ""}
	asset, errs = cfg.Configure()
	assert.NotEmpty(errs)
	assert.Empty(asset.Namespace)

	_, errs = cfg.Configure()
	assert.NotEmpty(errs)
}
