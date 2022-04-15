//go:build !windows
// +build !windows

// Package provides a unix-only system command execution
// and process limiting. Differs from sensu-go/command in
// that it does not assume a shell, and relies on
// context cancellation.
package v2

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/sensu/sensu-go/command"
	bytesutil "github.com/sensu/sensu-go/util/bytes"
	"github.com/sirupsen/logrus"
)

var (
	unixShellCommand []string = []string{"sh", "-c"}
	errEmptyCommand           = errors.New("execute requires a command")
)

type ExecutionRequest struct {
	// Command is the command to be executed.
	Command []string
	// Env ...
	Env []string
	// Input to provide the command via STDIN.
	Input string
}

func execute(ctx context.Context, execution ExecutionRequest) (*command.ExecutionResponse, error) {
	resp := &command.ExecutionResponse{}
	if len(execution.Command) == 0 {
		return nil, errEmptyCommand
	}
	executable, args := execution.Command[0], execution.Command[1:]
	logger := logrus.WithFields(logrus.Fields{"component": "commandv2"})
	var cmd *exec.Cmd
	cmd = exec.CommandContext(ctx, executable, args...)
	if len(execution.Env) > 0 {
		cmd.Env = execution.Env
	}

	var output bytesutil.SyncBuffer

	cmd.Stdout = &output
	cmd.Stderr = &output

	// If Input is specified, write to STDIN.
	if execution.Input != "" {
		cmd.Stdin = strings.NewReader(execution.Input)
	}

	started := time.Now()
	defer func() {
		resp.Duration = time.Since(started).Seconds()
	}()

	command.SetProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		// Something unexpected happened when attempting to
		// fork/exec, return immediately.
		return resp, err
	}
	waitCh := make(chan struct{})
	var err error
	go func() {
		err = cmd.Wait()
		close(waitCh)
	}()

	// Wait for the process to complete or the timer to trigger, whichever comes first.
	var killErr error
	select {
	case <-waitCh:
		resp.Output = output.String()
		if err != nil {
			// The command most likely return a non-zero exit status.
			if exitError, ok := err.(*exec.ExitError); ok {
				// Best effort to determine the exit status, this
				// should work on Linux, OSX, and Windows.
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					resp.Status = status.ExitStatus()
				} else {
					resp.Status = command.FallbackExitStatus
				}
			} else {
				resp.Status = command.FallbackExitStatus
			}
		} else {
			// Everything is A-OK.
			resp.Status = command.OKExitStatus
		}

	case <-ctx.Done():
		var killErrOutput string
		if killErr = command.KillProcess(cmd); killErr != nil {
			logger.WithError(killErr).Errorf("Execution timed out - Unable to TERM/KILL the process: #%d", cmd.Process.Pid)
			killErrOutput = fmt.Sprintf("Unable to TERM/KILL the process: #%d\n", cmd.Process.Pid)
		}
		resp.Output = fmt.Sprintf("%s%s%s", command.TimeoutOutput, killErrOutput, output.String())
		resp.Status = command.TimeoutExitStatus
	}

	return resp, nil
}

func ShellCommand(command string) []string {
	return append(unixShellCommand, command)
}
