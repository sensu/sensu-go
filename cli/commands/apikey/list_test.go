package apikey

import (
	"errors"
	"net/http"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/sensu/sensu-go/cli/commands/flags"
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
	assert.Regexp("api-keys", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.APIKey{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.APIKey)
			*resources = []corev2.APIKey{
				*corev2.FixtureAPIKey("my-api-key", "user1"),
				*corev2.FixtureAPIKey("my-other-api-key", "user2"),
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "my-api-key")
	assert.Contains(out, "my-other-api-key")
	assert.Nil(err)
	assert.NotContains(out, "==")
}

func TestListCommandRunEClosureWithAll(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.APIKey{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.APIKey)
			*resources = []corev2.APIKey{
				*corev2.FixtureAPIKey("my-api-key", "user1"),
			}
		},
	)

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

	apikey := corev2.FixtureAPIKey("my-api-key", "user1")
	client := cli.Client.(*client.MockClient)
	resources := []corev2.APIKey{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.APIKey)
			*resources = []corev2.APIKey{
				*apikey,
			}
		},
	)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Name")
	assert.Contains(out, "Username")
	assert.Contains(out, "Created At")
}

func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	resources := []corev2.APIKey{}
	client.On("List", mock.Anything, &resources, mock.Anything, mock.Anything).Return(errors.New("err"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotNil(err)
	assert.Equal("err", err.Error())
	assert.Empty(out)
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

func TestListCommandRunEClosureWithHeader(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	config := cli.Config.(*client.MockConfig)
	config.On("Format").Return("none")

	client := cli.Client.(*client.MockClient)
	var header http.Header
	resources := []corev2.APIKey{}
	client.On("List", mock.Anything, &resources, mock.Anything, &header).Return(nil).Run(
		func(args mock.Arguments) {
			resources := args[1].(*[]corev2.APIKey)
			*resources = []corev2.APIKey{}
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
