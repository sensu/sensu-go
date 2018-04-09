package extension

import (
	"errors"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestDeregisterCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeregisterCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("deregister", cmd.Use)
	assert.Regexp("extension", cmd.Short)
}

func TestDeregisterCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := DeregisterCommand(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Usage", out) // usage should print out
	assert.Error(err)
}

func TestDeregisterCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeregisterExtension", "my-extension", "default").Return(nil)

	cmd := DeregisterCommand(cli)
	out, err := test.RunCmd(cmd, []string{"my-extension"})

	assert.Regexp("OK", out)
	assert.Nil(err)
}

func TestDeregisterCommandRunEClosureWithServerErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("DeregisterExtension", "my-extension", "default").Return(errors.New("oh noes"))

	cmd := DeregisterCommand(cli)
	out, err := test.RunCmd(cmd, []string{"my-extension"})

	assert.Empty(out)
	assert.NotNil(err)
	assert.Equal("oh noes", err.Error())
}
