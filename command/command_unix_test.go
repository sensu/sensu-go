// +build !windows

package command

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExecuteUnix(t *testing.T) {
	// test that multiple commands can time out
	sleepMultiple := FakeCommand("sleep 10 && echo foo")
	sleepMultiple.Timeout = 1

	sleepMultipleExec, sleepMultipleErr := sleepMultiple.Execute(context.Background(), sleepMultiple)
	assert.Equal(t, nil, sleepMultipleErr)
	assert.Equal(t, "Execution timed out\n", testutil.CleanOutput(sleepMultipleExec.Output))
	assert.Equal(t, 2, sleepMultipleExec.Status)
	assert.NotEqual(t, 0, sleepMultipleExec.Duration)
}
