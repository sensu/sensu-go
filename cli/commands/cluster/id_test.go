package cluster

import (
	"errors"
	"os"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIDCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := IDCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("id", cmd.Use)
	assert.Regexp("sensu cluster id", cmd.Short)
}

func TestInfoCommandRunEClosure(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchClusterID").Return("foo", nil)

	cmd := IDCommand(cli)
	stdout.Start()
	_, err := test.RunCmd(cmd, []string{})
	stdout.Stop()
	require.NoError(t, err)
	assert.Regexp("foo", stdout.Output())
}

func TestInfoCommandWithArgs(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := IDCommand(cli)
	out, err := test.RunCmd(cmd, []string{"arg"})
	require.Error(t, err)

	assert.NotEmpty(out)
	assert.Contains(out, "Usage")
}

func TestInfoCommandRunEClosureWithErr(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)

	cli := test.NewCLI()
	client := cli.Client.(*client.MockClient)
	client.On("FetchClusterID").Return("", errors.New("err"))

	cmd := IDCommand(cli)
	stdout.Start()
	_, err := test.RunCmd(cmd, []string{})
	stdout.Stop()

	assert.NotNil(err)
	assert.Equal("err", err.Error())
	assert.Empty(stdout.Output())
}
