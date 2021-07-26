package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	test "github.com/sensu/sensu-go/cli/commands/testing"
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
	resources := []corev2.User{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.User)
			*resources = []corev2.User{
				*corev2.FixtureUser("one"),
				*corev2.FixtureUser("two"),
			}
		},
	)

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
	assert.NotContains(out, "==")
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.User{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(errors.New("fire"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Empty(out)
	assert.Error(err)
}

// TODO(ccressent): Combine all those output format tests into 1 test with
// subtests, to at least share the common initialization code.
func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLIWithValue("none")

	client := cli.Client.(*client.MockClient)
	resources := []corev2.User{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.User)
			*resources = []corev2.User{
				*corev2.FixtureUser("one"),
				*corev2.FixtureUser("two"),
			}
		},
	)

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

	testUsers := []corev2.User{
		*corev2.FixtureUser("user1"),
		*corev2.FixtureUser("user2"),
	}

	expected, err := json.Marshal(testUsers)
	if err != nil {
		t.Fatal(err)
	}

	client := cli.Client.(*client.MockClient)
	resources := []corev2.User{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.User)
			*resources = testUsers
		},
	)

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

	testUsers := []corev2.User{
		*corev2.FixtureUser("user1"),
		*corev2.FixtureUser("user2"),
	}

	client := cli.Client.(*client.MockClient)
	resources := []corev2.User{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.User)
			*resources = testUsers
		},
	)

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

func TestListCommandRunEClosureWithHeader(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	var header http.Header
	resources := []corev2.User{}
	client.On("List", mock.Anything, &resources, mock.Anything, &header).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.User)
			*resources = []corev2.User{}
			header := args[3].(*http.Header)
			*header = make(http.Header)
			header.Add(helpers.HeaderWarning, "E_TOO_MANY_ENTITIES")
		},
	)

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Nil(err)
	assert.Contains(out, "E_TOO_MANY_ENTITIES")
	assert.Contains(out, "==")
}
