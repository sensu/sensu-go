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

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
)

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

func TestExecute(t *testing.T) {
	// test that stdout can be read from
	echo := FakeCommand("echo", "foo")

	echoExec, echoErr := echo.Execute(context.Background(), echo)
	assert.Equal(t, nil, echoErr)
	assert.Equal(t, "foo\n", echoExec.Output)
	assert.Equal(t, 0, echoExec.Status)
	assert.NotEqual(t, 0, echoExec.Duration)

	// test that input can be passed to a command through stdin
	cat := FakeCommand("cat")
	cat.Input = "bar"

	catExec, catErr := cat.Execute(context.Background(), cat)
	assert.Equal(t, nil, catErr)
	assert.Equal(t, "bar", testutil.CleanOutput(catExec.Output))
	assert.Equal(t, 0, catExec.Status)
	assert.NotEqual(t, 0, catExec.Duration)

	// test that command exit codes can be read
	falseCmd := FakeCommand("false")

	falseExec, falseErr := falseCmd.Execute(context.Background(), falseCmd)
	assert.Equal(t, nil, falseErr)
	assert.Equal(t, "", testutil.CleanOutput(falseExec.Output))
	assert.Equal(t, 1, falseExec.Status)
	assert.NotEqual(t, 0, falseExec.Duration)

	// test that stderr can be read from
	outputs := FakeCommand("echo bar 1>&2")

	outputsExec, outputsErr := outputs.Execute(context.Background(), outputs)
	assert.Equal(t, nil, outputsErr)
	assert.Equal(t, "bar\n", testutil.CleanOutput(outputsExec.Output))
	assert.Equal(t, 0, outputsExec.Status)
	assert.NotEqual(t, 0, outputsExec.Duration)

	// test that commands can time out
	sleep := FakeCommand("sleep 10")
	sleep.Timeout = 1

	sleepExec, sleepErr := sleep.Execute(context.Background(), sleep)
	assert.Equal(t, nil, sleepErr)
	assert.Equal(t, "Execution timed out\n", testutil.CleanOutput(sleepExec.Output))
	assert.Equal(t, 2, sleepExec.Status)
	assert.NotEqual(t, 0, sleepExec.Duration)
}
