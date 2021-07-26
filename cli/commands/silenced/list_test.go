package silenced

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
	assert.Regexp("silenced entries", cmd.Short)
}

func TestListCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]corev2.Silenced{
		*corev2.FixtureSilenced("foo:bar"),
		*corev2.FixtureSilenced("bar:foo"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "json"))
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.Contains(out, "foo:bar")
	assert.Contains(out, "bar:foo")
	assert.Nil(err)
	assert.NotContains(out, "==")
}

func TestListCommandRunEClosureWithAll(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]corev2.Silenced{
		*corev2.FixtureSilenced("foo:bar"),
	}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set(flags.Format, "json"))
	require.NoError(t, cmd.Flags().Set(flags.AllNamespaces, "t"))
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out)
	assert.Nil(err)
}

func TestListCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)
	cli := test.NewCLIWithValue("none")

	silenced := corev2.FixtureSilenced("foo:bar")
	silenced.Reason = "justcause!"
	silenced.Creator = "eric"
	silenced.Check = "bar"
	silenced.Subscription = "foo"
	silenced.Namespace = "defaultnamespace"

	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]corev2.Silenced{*silenced}, nil)

	cmd := ListCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Name")            // heading
	assert.Contains(out, "Expiration")      // heading
	assert.Contains(out, "ExpireOnResolve") // heading
	assert.Contains(out, "Creator")         // heading
	assert.Contains(out, "Check")           // heading
	assert.Contains(out, "Reason")          // heading
	assert.Contains(out, "Subscription")    // heading
	assert.Contains(out, "Namespace")       // heading
	assert.Contains(out, "justcause!")
	assert.Contains(out, "foo:bar")
	assert.Contains(out, "eric")
	assert.Contains(out, "defaultnamespace")
	assert.Contains(out, "false")
	assert.Contains(out, "no expiration")
}

// Test to ensure silenced command list output does not escape alphanumeric chars
func TestListCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]corev2.Silenced{}, errors.New("my-err"))

	cmd := ListCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotNil(err)
	assert.Equal("my-err", err.Error())
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
	client.On("ListSilenceds", mock.Anything, mock.Anything, mock.Anything, mock.Anything, &header).Return([]corev2.Silenced{}, nil).Run(
		func(args mock.Arguments) {
			header := args[4].(*http.Header)
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
