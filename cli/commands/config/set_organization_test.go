package config

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/cli"
	clienttest "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestSetOrgCommand(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := SetOrgCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("set-organization", cmd.Use)
	assert.Regexp("Set organization", cmd.Short)
}

func TestSetOrgBadsArgs(t *testing.T) {
	assert := assert.New(t)

	cli := &cli.SensuCli{}
	cmd := SetOrgCommand(cli)

	// No args...
	out, err := test.RunCmd(cmd, []string{})
	assert.NotEmpty(out, "output should display help usage")
	assert.Error(err, "error should be returned")

	// Too many args...
	out, err = test.RunCmd(cmd, []string{"one", "two"})
	assert.NotEmpty(out, "output should display help usage")
	assert.Error(err, "error should be returned")
}

func TestSetOrgExec(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetOrgCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveOrganization", "default").Return(nil)

	out, err := test.RunCmd(cmd, []string{"default"})
	assert.Equal(out, "OK\n")
	assert.Nil(err, "Should not produce any errors")
}

func TestSetOrgWithWriteErr(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewMockCLI()
	cmd := SetOrgCommand(cli)

	config := cli.Config.(*clienttest.MockConfig)
	config.On("SaveOrganization", "default").Return(errors.New("blah"))

	out, err := test.RunCmd(cmd, []string{"default"})
	assert.Contains(out, "Unable to write")
	assert.Nil(err, "Should not return an error")
}
