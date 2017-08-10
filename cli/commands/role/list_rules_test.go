package role

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestListRulesCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("json")

	cmd := ListRulesCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("list-rules", cmd.Use)
	assert.Regexp("list rules", cmd.Short)
}

func TestListRulesCommandRunEWithError(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	client.On("FetchRole", "abc").Return(
		types.FixtureRole("abc", "default", "default"),
		errors.New("sadfa"),
	)

	cmd := ListRulesCommand(cli)
	out, err := test.RunCmd(cmd, []string{"abc"})

	assert.Empty(out)
	assert.Error(err)
}

func TestListRulesCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	client.On("FetchRole", "abc").Return(
		types.FixtureRole("abc", "default", "default"),
		nil,
	)

	cmd := ListRulesCommand(cli)
	out, err := test.RunCmd(cmd, []string{"abc"})

	assert.NotEmpty(out)
	assert.Contains(out, "Type")
	assert.Contains(out, "Org")
	assert.Contains(out, "Env")
	assert.Contains(out, "Permissions")
	assert.Nil(err)
}

func TestListRulesCommandRunEJSON(t *testing.T) {
	assert := assert.New(t)
	cli := newCLI()

	client := cli.Client.(*client.MockClient)
	client.On("FetchRole", "abc").Return(
		types.FixtureRole("abc", "default", "default"),
		nil,
	)

	cmd := ListRulesCommand(cli)
	out, err := test.RunCmd(cmd, []string{"abc"})

	assert.NotEmpty(out)
	assert.NoError(err)
}
