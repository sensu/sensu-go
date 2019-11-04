package apikey

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGrantCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := GrantCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("grant", cmd.Use)
	assert.Regexp("api-key", cmd.Short)
}

func TestGrantCommandWithoutArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("PostAPIKey", mock.Anything, mock.Anything).Return("", nil)

	cmd := GrantCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.NotEmpty(out)
	assert.NotNil(err)
}

func TestGrantCommandWithArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("PostAPIKey", mock.Anything, mock.Anything).Return("location", nil)

	cmd := GrantCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user1"})

	require.NoError(t, err)
	assert.Regexp("Created: location", out)
}
func TestGrantCommandServerError(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("PostAPIKey", mock.Anything, mock.Anything).Return("", errors.New("err"))

	cmd := GrantCommand(cli)
	out, err := test.RunCmd(cmd, []string{"user1"})

	assert.Empty(out)
	assert.Error(err)
	assert.Equal("err", err.Error())
}
