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

	echoExec, echoErr := ExecuteCommand(echo)
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", echoExec.Output)
	assert.Equal(t, 0, echoExec.Status)

	cat := &Execution{
		Command: "cat",
		Input:   "bar",
	}

	catExec, catErr := ExecuteCommand(cat)
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", catExec.Output)
	assert.Equal(t, 0, catExec.Status)

	falseCmd := &Execution{
		Command: "false",
	}

	falseExec, falseErr := ExecuteCommand(falseCmd)
	assert.Equal(t, nil, falseErr)
	assert.Equal(t, "", falseExec.Output)
	assert.Equal(t, 1, falseExec.Status)

	outputs := &Execution{
		Command: "echo foo && echo bar 1>&2",
	}

	outputsExec, outputsErr := ExecuteCommand(outputs)
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "foo\nbar\n", outputsExec.Output)
	assert.Equal(t, 0, outputsExec.Status)
}
