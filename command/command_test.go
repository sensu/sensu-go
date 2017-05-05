// Package command provides system command execution.
package command

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/util"
	"github.com/stretchr/testify/assert"
)

func fakeCommand(command string, args ...string) *Execution {
	cs := []string{os.Args[0], "-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmdStr := strings.Join(cs, " ")
	trimmedCmd := strings.Trim(cmdStr, " ")
	env := []string{"GO_WANT_HELPER_PROCESS=1"}

	execution := &Execution{
		Command: trimmedCmd,
		Env:     env,
	}

	return execution
}

func TestHelperProcess(t *testing.T) {
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

func TestExecuteCommand(t *testing.T) {
	// test that stdout can be read from
	echo := fakeCommand("echo", "foo")

	echoExec, echoErr := ExecuteCommand(context.Background(), echo)
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", echoExec.Output)
	assert.Equal(t, 0, echoExec.Status)
	assert.NotEqual(t, 0, echoExec.Duration)

	// test that input can be passed to a command through stdin
	cat := fakeCommand("cat")
	cat.Input = "bar"

	catExec, catErr := ExecuteCommand(context.Background(), cat)
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", util.CleanOutput(catExec.Output))
	assert.Equal(t, 0, catExec.Status)
	assert.NotEqual(t, 0, catExec.Duration)

	// test that command exit codes can be read
	falseCmd := fakeCommand("false")

	falseExec, falseErr := ExecuteCommand(context.Background(), falseCmd)
	assert.Equal(t, nil, falseErr)
	assert.Equal(t, "", util.CleanOutput(falseExec.Output))
	assert.Equal(t, 1, falseExec.Status)
	assert.NotEqual(t, 0, falseExec.Duration)

	// test that stderr can be read from
	outputs := fakeCommand("echo bar 1>&2")

	outputsExec, outputsErr := ExecuteCommand(context.Background(), outputs)
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "bar\n", util.CleanOutput(outputsExec.Output))
	assert.Equal(t, 0, outputsExec.Status)
	assert.NotEqual(t, 0, outputsExec.Duration)

	// test that commands can time out
	sleep := fakeCommand("sleep 10")
	sleep.Timeout = 1

	sleepExec, sleepErr := ExecuteCommand(context.Background(), sleep)
	assert.Equal(t, nil, sleepErr)
	assert.Equal(t, "Execution timed out\n", util.CleanOutput(sleepExec.Output))
	assert.Equal(t, 2, sleepExec.Status)
	assert.NotEqual(t, 0, sleepExec.Duration)
}
