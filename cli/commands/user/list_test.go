package user

import (
	"encoding/json"
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
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
	assert.Regexp("user", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListUsers", mock.Anything).Return([]types.User{
		*types.FixtureUser("one"),
		*types.FixtureUser("two"),
	}, nil)

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListUsers", mock.Anything).Return([]types.User{}, errors.New("fire"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}

// TODO(ccressent): Combine all those output format tests into 1 test with
// subtests, to at least share the common initialization code.
func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	client := cli.Client.(*client.MockClient)
	client.On("ListUsers", mock.Anything).Return([]types.User{
		*types.FixtureUser("one"),
		*types.FixtureUser("two"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "Username")
	assert.Contains(out, "Groups")
	assert.Contains(out, "Enabled")
	assert.Contains(out, "one")
	assert.Contains(out, "two")
	assert.Contains(out, "true")
	assert.NoError(err)
}

func TestListCommandRunEClosureWithJSONOutput(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	testUsers := []types.User{
		*types.FixtureUser("user1"),
		*types.FixtureUser("user2"),
	}

	expected, err := json.Marshal(testUsers)
	if err != nil {
		t.Fatal(err)
	}

	client := cli.Client.(*client.MockClient)
	client.On("ListUsers", mock.Anything).Return(testUsers, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NoError(err)
	assert.NotEmpty(out)
	assert.JSONEq(string(expected), out)
}

func TestListCommandRunEClosureWithWrappedJSONOutput(t *testing.T) {
	// User does not meet the Resource interface (no ObjectMeta), so the
	// "wrapped-json" output for it should be the exact same as the "json"
	// output.
	TestListCommandRunEClosureWithJSONOutput(t)
}

func TestListCommandRunEClosureWithYAMLOutput(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLI()

	testUsers := []types.User{
		*types.FixtureUser("user1"),
		*types.FixtureUser("user2"),
	}

	client := cli.Client.(*client.MockClient)
	client.On("ListUsers", mock.Anything).Return(testUsers, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "yaml"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NoError(err)
	assert.NotEmpty(out)
	assert.Contains(out, "username")
	assert.Contains(out, "groups")
	assert.Contains(out, "disabled")
	assert.Contains(out, "user1")
	assert.Contains(out, "user2")
	assert.Contains(out, "false")
}
