package entity

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
	assert.Regexp("entities", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEntities", mock.Anything).Return([]types.Entity{
		*types.FixtureEntity("name-one"),
		*types.FixtureEntity("name-two"),
	}, nil)

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "name-one")
	assert.Contains(out, "name-two")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithAllOrgs(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEntities", "*").Return([]types.Entity{
		*types.FixtureEntity("name-two"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set(flags.AllOrgs, "t")
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEntities", mock.Anything).Return([]types.Entity{
		*types.FixtureEntity("name-one"),
		*types.FixtureEntity("name-two"),
	}, nil)

	cmd := ListCommand(cli)
	cmd.Flags().Set(flags.Format, "none")

	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "ID")
	assert.Contains(out, "OS")
	assert.Contains(out, "Subscriptions")
	assert.Contains(out, "Last Seen")
	assert.Nil(err)
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := newCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListEntities", mock.Anything).Return([]types.Entity{}, errors.New("my-err"))

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
