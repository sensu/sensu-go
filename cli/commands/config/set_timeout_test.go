package config

import (
	"errors"
	"testing"
	"time"

	"github.com/sensu/sensu-go/cli"
	clienttest "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestSetTimeoutCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := SetTimeoutCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set-timeout", cmd.Use)
	assert.Regexp("Set timeout", cmd.Short)
}

func TestSetTimeoutBadsArgs(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := SetTimeoutCommand(cli)

	// No args...
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out, "output should display help usage")
	assert.Error(err, "error should be returned")

	// Too many args...
	out, err = test.RunCmd(cmd, []string{"one", "two"})
	assert.NotEmpty(out, "output should display help usage")
	assert.Error(err, "error should be returned")
}

func TestSetTimeoutExec(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetTimeoutCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveTimeout", 15*time.Second).Return(nil)

	out, err := test.RunCmd(cmd, []string{"15s"})
	assert.Equal(out, "Updated\n")
	assert.Nil(err, "Should not produce any errors")
}

func TestSetTimeoutWithWriteErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetTimeoutCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveTimeout", 15*time.Second).Return(errors.New("blah"))

	out, err := test.RunCmd(cmd, []string{"15s"})
	assert.Contains(out, "Unable to write")
	assert.Nil(err, "Should not return an error")
}
