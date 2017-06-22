package asset

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
	cmd.Flags().Set("url", "http://lol")
	cmd.Flags().Set("sha512", "12345qwerty")
	out, err := test.RunCmd(cmd, []string{"ruby22"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestCreateCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("CreateAsset", mock.AnythingOfType("*types.Asset")).Return(errors.New("whoops"))

	cmd := CreateCommand(cli)
	cmd.Flags().Set("sha512", "12345qwerty")
	cmd.Flags().Set("url", "http://lol")
	out, err := test.RunCmd(cmd, []string{"ruby22"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("whoops", err.Error())
}

func TestCreateExectorBadURLGiven(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()
	cmd := CreateCommand(cli)

	cmd.Flags().Set("url", "my-bad-bad-url-boy")
	exec := &CreateExecutor{Client: cli.Client}

	err := exec.Run(cmd, []string{"ruby22"})
	assert.Error(err)
}

type ConfigureAssetSuite struct {
	suite.Suite
	flags *pflag.FlagSet
}

func (suite *ConfigureAssetSuite) SetupTest() {
	flags := &pflag.FlagSet{}
	flags.StringSlice("metadata", []string{}, "")
	flags.String("sha512", "12345qwerty", "")
	flags.StringSlice("filter", []string{}, "")
	flags.String("url", "http://lol", "")
	suite.flags = flags
}

// Too many args
func (suite *ConfigureAssetSuite) TestTooManyArgs() {
	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{"one", "too many"},
		Org:   "default",
	}

	asset, errs := cfg.Configure()
	suite.NotEmpty(errs, "Error claiming an over abundance of flags is present")
	suite.Empty(asset.Name, "Name should not be set")
}

// Empty org
func (suite *ConfigureAssetSuite) TestEmptyOrg() {
	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{"ruby22"},
		Org:   "",
	}

	asset, errs := cfg.Configure()
	suite.NotEmpty(errs, "Error should be present for missing org")
	suite.Empty(asset.Organization, "Organization should not be set")
}

// No args
func (suite *ConfigureAssetSuite) TestMissingArgs() {
	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{},
		Org:   "default",
	}

	asset, errs := cfg.Configure()
	suite.NotEmpty(errs, "Should contain error for missing name")
	suite.Empty(asset.Name, "Should not be set")
}

// Given name
func (suite *ConfigureAssetSuite) TestGivenValidName() {
	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{"ruby22"},
		Org:   "default",
	}

	asset, errs := cfg.Configure()
	suite.Empty(errs)
	suite.Equal("ruby22", asset.Name)
}

// Valid Metadata
func (suite *ConfigureAssetSuite) TestValidMetadata() {
	suite.flags.Set("metadata", "One: Two")
	suite.flags.Set("metadata", "  Three : Four ")

	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{"ruby22"},
		Org:   "default",
	}

	asset, errs := cfg.Configure()
	suite.Empty(errs)
	suite.NotEmpty(asset.Metadata, "Metadata field should have been set")
	suite.Equal("Two", asset.Metadata["One"], "Metadata param 'one' is set")
}

// Bad Metadata
func (suite *ConfigureAssetSuite) TestBadMetadata() {
	suite.flags.Set("metadata", "Five- Six")

	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{"ruby22"},
		Org:   "default",
	}

	asset, errs := cfg.Configure()
	suite.NotEmpty(errs, "Contains error regarding metadata format")
	suite.Empty(asset.Metadata, "Metadata is not set")
}

// Valid filters
func (suite *ConfigureAssetSuite) TestValidFilters() {
	suite.flags.Set("filter", "entity.System.OS = 'meowmix'")

	cfg := ConfigureAsset{
		Flags: suite.flags,
		Args:  []string{"ruby22"},
		Org:   "default",
	}

	asset, errs := cfg.Configure()
	suite.Empty(errs)
	suite.NotEmpty(asset.Filters, "Filters should have been set")
}

func TestRunSuites(t *testing.T) {
	suite.Run(t, new(ConfigureAssetSuite))
}
