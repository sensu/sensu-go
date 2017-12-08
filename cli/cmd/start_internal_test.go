package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
)

// Borrowed from testify.suite package
// https://github.com/stretchr/testify/blob/v1.1.4/suite/suite_test.go
type StdoutCapture struct {
	oldStdout *os.File
	readPipe  *os.File
}

func (sc *StdoutCapture) StartCapture() {
	sc.oldStdout = os.Stdout
	sc.readPipe, os.Stdout, _ = os.Pipe()
}

func (sc *StdoutCapture) StopCapture() (string, error) {
	if sc.oldStdout == nil || sc.readPipe == nil {
		return "", errors.New("StartCapture not called before StopCapture")
	}
	if err := os.Stdout.Close(); err != nil {
		return "", err
	}
	os.Stdout = sc.oldStdout
	bytes, err := ioutil.ReadAll(sc.readPipe)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

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
