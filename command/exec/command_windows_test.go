//go:build windows
// +build windows

package exec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecuteWindows(t *testing.T) {
	outputBuf := strings.Builder{}
	timeoutRequest := ExecutionRequest{
		Command: ShellCommand("PowerShell.exe Write-Host \"start sleep\"; Start-Sleep -s 20 ; Write-Host \"sleep done\""),
		Stdout: &outputBuf
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()


	sleepErr := timeoutRequest.Execute(ctx)
	timeoutErr, ok := sleepErr .(TimeoutError)
	assert.Truef(t, ok, "expected timeout error")
	// within 20% of the 1 second timeout
	duration := timeoutErr.Timeout().Seconds()
	assert.Less(t, 0.8, duration)
	assert.Greater(t, 1.2, duration)
}
