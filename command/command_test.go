// Package command provides system command execution.
package command

import (
	"context"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/testing/util"
	"github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
	echo := &Execution{
		Command: "echo foo",
	}

	echoExec, echoErr := ExecuteCommand(context.Background(), echo)
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", util.CleanOutput(echoExec.Output))
	assert.Equal(t, 0, echoExec.Status)
	assert.NotEqual(t, 0, echoExec.Duration)

	catPath := util.CommandPath("cat")

	cat := &Execution{
		Command: catPath,
		Input:   "bar",
	}

	catExec, catErr := ExecuteCommand(context.Background(), cat)
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", util.CleanOutput(catExec.Output))
	assert.Equal(t, 0, catExec.Status)
	assert.NotEqual(t, 0, catExec.Duration)

	falsePath := util.CommandPath("false")

	falseCmd := &Execution{
		Command: falsePath,
	}

	falseExec, falseErr := ExecuteCommand(context.Background(), falseCmd)
	assert.Equal(t, nil, falseErr)
	assert.Equal(t, "", util.CleanOutput(falseExec.Output))
	assert.Equal(t, 1, falseExec.Status)
	assert.NotEqual(t, 0, falseExec.Duration)

	outputs := &Execution{
		Command: "(echo foo) && (echo bar) 1>&2",
	}

	outputsExec, outputsErr := ExecuteCommand(context.Background(), outputs)
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "foo\nbar\n", util.CleanOutput(outputsExec.Output))
	assert.Equal(t, 0, outputsExec.Status)
	assert.NotEqual(t, 0, outputsExec.Duration)

	sleepPath := util.CommandPath("sleep")

	sleep := &Execution{
		Command: fmt.Sprintf("%s 10", sleepPath),
		Timeout: 1,
	}

	sleepExec, sleepErr := ExecuteCommand(context.Background(), sleep)
	assert.Equal(t, nil, sleepErr)
	assert.Equal(t, "Execution timed out\n", util.CleanOutput(sleepExec.Output))
	assert.Equal(t, 2, sleepExec.Status)
	assert.NotEqual(t, 0, sleepExec.Duration)
}
