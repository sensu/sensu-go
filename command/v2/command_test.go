//go:build !windows
// +build !windows

package v2

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHelperProcessV2(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	command := strings.Join(os.Args[3:], " ")

	stdin, _ := ioutil.ReadAll(os.Stdin)

	hasArgs := len(os.Args) > 4
	argStr := ""
	if hasArgs {
		argStr = strings.Join(os.Args[4:], " ")
	}

	switch command {
	case "cat":
		fmt.Fprintf(os.Stdout, "%s", stdin)
	case "echo foo":
		fmt.Fprintln(os.Stdout, argStr)
	case "echo bar":
		fmt.Fprintln(os.Stderr, argStr)
	case "false":
		os.Exit(1)
	case "sleep 10":
		time.Sleep(10 * time.Second)
	}
	os.Exit(0)
}

func TestExecute(t *testing.T) {
	// test that stdout can be read from
	echo := FakeCommand("echo", "foo")

	echoOut := strings.Builder{}
	echo.Stdout = &echoOut
	echoErr := echo.Execute(context.Background())
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", echoOut.String())

	// test that input can be passed to a command through stdin
	cat := FakeCommand("cat")
	catOut := strings.Builder{}
	cat.Stdin = strings.NewReader("bar")
	cat.Stdout = &catOut

	catErr := cat.Execute(context.Background())
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", testutil.CleanOutput(catOut.String()))

	// test that command exit codes can be read
	falseCmd := FakeCommand("false")
	falseOut := strings.Builder{}
	falseCmd.Stdout, falseCmd.Stderr = &falseOut, &falseOut

	falseErr := falseCmd.Execute(context.Background())
	exitErr, ok := falseErr.(ExitError)
	assert.Truef(t, ok, "expected exit code error")
	assert.Equal(t, "", testutil.CleanOutput(falseOut.String()))
	assert.Equal(t, 1, exitErr.ExitStatus())

	// test that stderr can be read from
	outputCmd := FakeCommand("echo", "bar")
	outputOut := strings.Builder{}
	outputCmd.Stdout, outputCmd.Stderr = &outputOut, &outputOut

	outputsErr := outputCmd.Execute(context.Background())
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "bar\n", testutil.CleanOutput(outputOut.String()))

	// test that commands can time out
	sleep := FakeCommand("sleep 10")
	sleepOut := strings.Builder{}
	sleep.Stdout, sleep.Stderr = &sleepOut, &sleepOut

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	sleepErr := sleep.Execute(ctx)
	timeoutErr, ok := sleepErr.(TimeoutError)
	assert.Truef(t, ok, "expected timeout error")
	assert.Equal(t, "", testutil.CleanOutput(sleepOut.String()))
	assert.NotEqual(t, 0, timeoutErr.Timeout())
}

// FakeCommand takes a command and (optionally) command args and will execute
// the TestHelperProcess test within the package FakeCommand is called from.
func FakeCommand(command string, args ...string) ExecutionRequest {
	env := []string{"GO_WANT_HELPER_PROCESS=1"}

	execution := ExecutionRequest{
		Command: []string{
			os.Args[0],
			"-test.run=TestHelperProcessV2",
			"--",
			command,
		},
		Stdin: strings.NewReader("bar"),
		Env:   env,
	}

	execution.Command = append(execution.Command, args...)
	return execution
}

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
