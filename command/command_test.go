// Package command provides system command execution.
package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	echo := &Execution{
		Command: "echo foo",
	}

	echoExec, echoErr := ExecuteCommand(context.Background(), echo)
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", echoExec.Output)
	assert.Equal(t, 0, echoExec.Status)
	assert.NotEqual(t, 0, echoExec.Duration)

	cat := &Execution{
		Command: "cat",
		Input:   "bar",
	}

	catExec, catErr := ExecuteCommand(context.Background(), cat)
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", catExec.Output)
	assert.Equal(t, 0, catExec.Status)

	falseCmd := &Execution{
		Command: "false",
	}

	falseExec, falseErr := ExecuteCommand(context.Background(), falseCmd)
	assert.Equal(t, nil, falseErr)
	assert.Equal(t, "", falseExec.Output)
	assert.Equal(t, 1, falseExec.Status)

	outputs := &Execution{
		Command: "echo foo && echo bar 1>&2",
	}

	outputsExec, outputsErr := ExecuteCommand(context.Background(), outputs)
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "foo\nbar\n", outputsExec.Output)
	assert.Equal(t, 0, outputsExec.Status)

	sleep := &Execution{
		Command: "sleep 10",
		Timeout: 1,
	}

	sleepExec, sleepErr := ExecuteCommand(context.Background(), sleep)
	assert.Equal(t, nil, sleepErr)
	assert.Equal(t, "Execution timed out\n", sleepExec.Output)
	assert.Equal(t, 2, sleepExec.Status)
}
