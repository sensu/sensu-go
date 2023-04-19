package role

import (
	"errors"
	"testing"

	v2 "github.com/sensu/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestInfoCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("json")

	cmd := InfoCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("info", cmd.Use)
	assert.Regexp("show detailed information about a role", cmd.Short)
}

func TestInfoCommandRunEWithError(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	client.On("FetchRole", "abc").Return(
		v2.FixtureRole("abc", "default"),
		errors.New("sadfa"),
	)

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"abc"})

	assert.Empty(out)
	assert.Error(err)
}

func TestInfoCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewMockCLI()

	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	client.On("FetchRole", "abc").Return(
		v2.FixtureRole("abc", "default"),
		nil,
	)

	cmd := InfoCommand(cli)
	_, err := test.RunCmd(cmd, []string{"abc"})

	//assert.NotEmpty(out)
	assert.Nil(err)
}

func TestInfoCommandRunEJSON(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	client := cli.Client.(*client.MockClient)
	client.On("FetchRole", "abc").Return(
		v2.FixtureRole("abc", "default"),
		nil,
	)

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"abc"})

	assert.NotEmpty(out)
	assert.NoError(err)
}
