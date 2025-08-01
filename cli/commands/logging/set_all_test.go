package logging

import (
	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSetLogLevelForAllModulesForModuleCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetLogLevelForAllModules(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set-all", cmd.Use)
	assert.Regexp("set same loglevel for all modules", cmd.Short)
}

func TestSetLogLevelForAllModulesCommandRunEClosureWithoutName(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetLogLevelForAllModules(cli)
	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("--level", err) // usage should print out
	assert.Empty(out)
}

func TestSetLogLevelForAllModulesCommandRunEClosureWithFlags(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	client := cli.Client.(*client.MockClient)
	client.On("SetLogLevelAllModules", mock.Anything).Return(nil)

	cmd := SetLogLevelForAllModules(cli)
	cmd.Flags().Set("level", "debug")

	out, err := test.RunCmd(cmd, []string{})

	assert.Regexp("Log level", out)
	assert.Nil(err)
}
