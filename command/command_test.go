// Package command provides system command execution.
package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	echo := &Execution{
		Command: "echo foo",
	}

	cmd, err := ExecuteCommand(echo)
	assert.Equal(t, nil, err)
	assert.Equal(t, "foo\n", cmd.Output)
	assert.Equal(t, 0, cmd.Status)
}
