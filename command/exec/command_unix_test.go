//go:build !windows
// +build !windows

package exec

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExecuteUnix(t *testing.T) {
	// test that multiple commands can time out
	outputBuf := strings.Builder{}
	sleepMultiple := ExecutionRequest{
		Command: ShellCommand("echo $TEST_EXECUTE_UNIX_VAR && sleep 10"),
		Env:     []string{"TEST_EXECUTE_UNIX_VAR=test-echo-output"},
		Stdout:  &outputBuf,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sleepMultipleErr := sleepMultiple.Execute(ctx)
	timeoutErr, ok := sleepMultipleErr.(TimeoutError)
	assert.Truef(t, ok, "expected timeout error")

	assert.Equal(t, "test-echo-output\n", testutil.CleanOutput(outputBuf.String()))
	assert.NotEqual(t, 0, timeoutErr.Timeout())
	// within 20% of the 1 second timeout
	duration := timeoutErr.Timeout().Seconds()
	assert.Less(t, 0.8, duration)
	assert.Greater(t, 1.2, duration)

	// test Polite Termination workflow
	outputBuf = strings.Builder{}
	sleepMultiplePolite := ExecutionRequest{
		Command: ShellCommand(`trap "" SIGTERM; sleep 15`),
		Stdout:  &outputBuf,
		Timeout: TimeoutPolitelyTerminate(time.Now().Add(time.Millisecond*400), 2, time.Millisecond*200),
	}
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sleepMultipleErr = sleepMultiplePolite.Execute(ctx)
	timeoutErr, ok = sleepMultipleErr.(TimeoutError)
	assert.Truef(t, ok, "expected timeout error")

	assert.Equal(t, "", testutil.CleanOutput(outputBuf.String()))
	assert.NotEqual(t, 0, timeoutErr.Timeout())
	// within 10% of the expected 1 second (400ms initial timeout, + 2 SIGTERMs sent 200ms apart + SIGKILL after final 200ms)
	duration = timeoutErr.Timeout().Seconds()
	assert.Less(t, 0.9, duration)
	assert.Greater(t, 1.1, duration)

}
