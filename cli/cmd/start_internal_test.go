package main

import (
	"os"
	"testing"

	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

func TestConfigureVersionCmd(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)
	cmd := newVersionCommand()
	assert.NotNil(cmd, "Returns a Command instance")
	assert.Equal("version", cmd.Use, "Configures the name")

	// Run command w/o any flags
	stdout.Start()
	cmd.Run(cmd, []string{})
	stdout.Stop()
	assert.Regexp("sensu-ctl version", stdout.Output())
}

func TestConfigureRootCmd(t *testing.T) {
	assert := assert.New(t)
	stdout := test.NewFileCapture(&os.Stdout)
	cmd := configureRootCmd()

	assert.NotNil(cmd, "Returns a Command instance")
	assert.Equal("sensuctl", cmd.Use, "Configures the name")

	// Run command w/o any flags
	stdout.Start()
	cmd.Run(cmd, []string{})
	stdout.Stop()
	assert.Regexp("Usage:", stdout.Output())
}
