package extension

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RegisterCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("register", cmd.Use)
	assert.Regexp("extensions", cmd.Short)
}

func TestRegisterCommandRunEClosureWithoutFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := RegisterCommand(cli)
	_, err := test.RunCmd(cmd, []string{"my-extensions"})

	assert.Error(err)
}

func TestRegisterCommandRunEClosureWithAllFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RegisterExtension", mock.AnythingOfType("*types.Extension")).Return(nil)

	cmd := RegisterCommand(cli)
	out, err := test.RunCmd(cmd, []string{"frobber", "http://localhost"})
	require.NoError(t, err)

	assert.Regexp("OK", out)
}

func TestRegisterCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("RegisterExtension", mock.AnythingOfType("*types.Extension")).Return(errors.New("whoops"))

	cmd := RegisterCommand(cli)
	out, err := test.RunCmd(cmd, []string{"frobber", "http://localhost"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("whoops", err.Error())
}
