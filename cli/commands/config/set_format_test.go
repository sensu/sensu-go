package config

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/cli"
	clienttest "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestSetFormatCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := SetFormatCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set-format", cmd.Use)
	assert.Regexp("Set format", cmd.Short)
}

func TestSetFormatBadsArgs(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := SetFormatCommand(cli)

	// No args...
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out, "output should display help usage")
	assert.Error(err, "error should be returned")

	// Too many args...
	out, err = test.RunCmd(cmd, []string{"one", "two"})
	assert.NotEmpty(out, "output should display help usage")
	assert.Error(err, "error should be returned")
}

func TestSetFormatExec(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetFormatCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveFormat", "json").Return(nil)

	out, err := test.RunCmd(cmd, []string{"json"})
	assert.Equal(out, "OK\n")
	assert.Nil(err, "Should not produce any errors")
}

func TestSetFormatWithWriteErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetFormatCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveFormat", "json").Return(errors.New("blah"))

	out, err := test.RunCmd(cmd, []string{"json"})
	assert.Contains(out, "Unable to write")
	assert.Nil(err, "Should not return an error")
}
