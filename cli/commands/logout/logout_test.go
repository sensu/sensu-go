package logout

import (
	"fmt"
	"testing"

	clienttest "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogout(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := Command(cli)

	client := cli.Client.(*clienttest.MockClient)
	client.On("Logout", "bar").Return(nil)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveTokens", mock.AnythingOfType("*types.Tokens")).Return(nil)
	tokens := types.FixtureTokens("foo", "bar")
	config.On("Tokens").Return(tokens)

	out, err := test.RunCmd(cmd, []string{})
	assert.Regexp(t, "logout", out)
	assert.Nil(t, err)
}

func TestLogoutServerError(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := Command(cli)

	client := cli.Client.(*clienttest.MockClient)
	client.On("Logout", "bar").Return(fmt.Errorf("error"))

	config := cli.Config.(*clienttest.MockConfig)
	tokens := types.FixtureTokens("foo", "bar")
	config.On("Tokens").Return(tokens)

	out, err := test.RunCmd(cmd, []string{"bar"})
	// No error, print help usage
	assert.NotEmpty(t, out)
	assert.Error(t, err)
}

func TestLogoutServerConfigFile(t *testing.T) {
	cli := test.NewMockCLI()
	cmd := Command(cli)

	client := cli.Client.(*clienttest.MockClient)
	client.On("Logout", "bar").Return(nil)

	config := cli.Config.(*clienttest.MockConfig)
	tokens := types.FixtureTokens("foo", "bar")
	config.On("SaveTokens", mock.AnythingOfType("*types.Tokens")).Return(fmt.Errorf("error"))
	config.On("Tokens").Return(tokens)

	out, err := test.RunCmd(cmd, []string{"bar"})
	// Print usage
	assert.NotEmpty(t, out)
	assert.Error(t, err)
}
