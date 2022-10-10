package apikey

import (
	"errors"
	"testing"

	corev2 "github.com/sensu/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("info", cmd.Use)
	assert.Regexp("api-key", cmd.Short)
}

func TestInfoCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: "my-api-key",
		},
	}
	client.On("Get", apikey.URIPath(), apikey).Return(nil)

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"my-api-key"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "my-api-key")
}

func TestInfoCommandRunMissingArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{})
	require.Error(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Usage")
}

func TestInfoCommandRunEClosureWithTable(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: "my-api-key",
		},
	}
	client.On("Get", apikey.URIPath(), apikey).Return(nil)

	cmd := InfoCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "tabular"))

	out, err := test.RunCmd(cmd, []string{"my-api-key"})
	require.NoError(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Name")
	assert.Contains(out, "Username")
	assert.Contains(out, "Created At")
}

func TestInfoCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: "my-api-key",
		},
	}
	client.On("Get", apikey.URIPath(), apikey).Return(errors.New("err"))

	cmd := InfoCommand(cli)
	out, err := test.RunCmd(cmd, []string{"my-api-key"})

	assert.NotNil(err)
	assert.Equal("err", err.Error())
	assert.Empty(out)
}
