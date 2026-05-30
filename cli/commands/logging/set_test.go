package logging

import (
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSetLogLevelForModuleCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetLogLevelForModule(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set", cmd.Use)
	assert.Regexp("set loglevel", cmd.Short)
}

func TestSetLogLevelForModuleCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetLogLevelForModule(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("--module", err) // usage should print out
	assert.Empty(out)
}

func TestSetLogLevelForModuleCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("SetLogLevel", mock.Anything).Return(nil)

	cmd := SetLogLevelForModule(cli)
	cmd.Flags().Set("level", "debug")
	cmd.Flags().Set("module", "apid")

	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Log level", out)
	assert.Nil(err)
}
