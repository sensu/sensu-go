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

	echoExec, echoErr := execute(context.Background(), echo)
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", echoExec.Output)
	assert.Equal(t, 0, echoExec.Status)
	assert.NotEqual(t, 0, echoExec.Duration)

	// test that input can be passed to a command through stdin
	cat := FakeCommand("cat")
	cat.Input = "bar"

	catExec, catErr := execute(context.Background(), cat)
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", testutil.CleanOutput(catExec.Output))
	assert.Equal(t, 0, catExec.Status)
	assert.NotEqual(t, 0, catExec.Duration)

	// test that command exit codes can be read
	falseCmd := FakeCommand("false")

	falseExec, falseErr := execute(context.Background(), falseCmd)
	assert.Equal(t, nil, falseErr)
	assert.Equal(t, "", testutil.CleanOutput(falseExec.Output))
	assert.Equal(t, 1, falseExec.Status)
	assert.NotEqual(t, 0, falseExec.Duration)

	// test that stderr can be read from
	outputs := FakeCommand("echo", "bar")

	outputsExec, outputsErr := execute(context.Background(), outputs)
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "bar\n", testutil.CleanOutput(outputsExec.Output))
	assert.Equal(t, 0, outputsExec.Status)
	assert.NotEqual(t, 0, outputsExec.Duration)

	// test that commands can time out
	sleep := FakeCommand("sleep 10")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	sleepExec, sleepErr := execute(ctx, sleep)
	assert.Equal(t, nil, sleepErr)
	assert.Equal(t, "Execution timed out\n", testutil.CleanOutput(sleepExec.Output))
	assert.Equal(t, 2, sleepExec.Status)
	assert.NotEqual(t, 0, sleepExec.Duration)
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
		Input: "bar",
		Env:   env,
	}

	execution.Command = append(execution.Command, args...)
	return execution
}

func TestExecuteUnix(t *testing.T) {
	// test that multiple commands can time out
	sleepMultiple := ExecutionRequest{
		Command: ShellCommand("echo $TEST_EXECUTE_UNIX_VAR && sleep 10"),
		Env:     []string{"TEST_EXECUTE_UNIX_VAR=test-echo-output"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	sleepMultipleExec, sleepMultipleErr := execute(ctx, sleepMultiple)
	assert.Equal(t, nil, sleepMultipleErr)
	assert.Equal(t, "Execution timed out\ntest-echo-output\n", testutil.CleanOutput(sleepMultipleExec.Output))
	assert.Equal(t, 2, sleepMultipleExec.Status)
	assert.NotEqual(t, 0, sleepMultipleExec.Duration)
	// within 20% of the 1 second timeout
	assert.Less(t, 0.8, sleepMultipleExec.Duration)
	assert.Greater(t, 1.2, sleepMultipleExec.Duration)
}
