//go:build windows
// +build windows

package command

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExecuteWindows(t *testing.T) {
	// test that commands can time out on windows
	timeoutRequest := ExecutionRequest{
		Command: "PowerShell.exe Write-Host \"start sleep\"; Start-Sleep -s 20 ; Write-Host \"sleep done\"",
		Timeout: 10,
		Name:    "sleep-test",
	}
	timeout := NewExecutor()

	sleepExec, sleepErr := timeout.Execute(context.Background(), timeoutRequest)
	assert.Equal(t, nil, sleepErr)
	cleanOutput := testutil.CleanOutput(sleepExec.Output)
	assert.Contains(t, cleanOutput, "Execution timed out\n")
	assert.Contains(t, cleanOutput, "start sleep")
	assert.NotContains(t, "sleep done", cleanOutput)
	assert.Equal(t, TimeoutExitStatus, sleepExec.Status)
	assert.NotEqual(t, 1, sleepExec.Duration)
}
