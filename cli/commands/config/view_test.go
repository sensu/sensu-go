package config

import (
	"testing"

	clienttest "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestViewCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ViewCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("APIUrl").Return("http://127.0.0.1:8080")
	config.On("Format").Return("none")

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("view", cmd.Use)
	assert.Regexp("Display active configuration", cmd.Short)
}

func TestViewExec(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := ViewCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("APIUrl").Return("http://127.0.0.1:8080")
	config.On("Format").Return("none")

	out, err := test.RunCmd(cmd, []string{"default"})
	assert.Regexp("Active Configuration", out)
	assert.Nil(err, "Should not produce any errors")
}
