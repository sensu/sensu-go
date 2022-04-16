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
	"io"
	"os/exec"
	"syscall"
	"time"

	"github.com/sensu/sensu-go/command"
	"github.com/sirupsen/logrus"
)

var (
	unixShellCommand []string = []string{"sh", "-c"}
	errEmptyCommand           = errors.New("execute requires a command")
)

type ExecutionRequest struct {
	// Command is the command to be executed.
	Command []string
	// Env environment variables in "foo=bar" format
	Env []string
	// Stdin
	Stdin io.Reader
	// Stdout
	Stdout io.Writer
	// Stderr
	Stderr io.Writer
}

// ExecutionTimeout is returned when a command
// is interrupted by context cancellation
type ExecutionTimeout interface {
	Timeout() time.Duration
}

type timeoutError struct {
	time.Duration
	Wrapped error
}

func (t timeoutError) Error() string {
	if t.Wrapped != nil {
		return t.Wrapped.Error()
	}
	return fmt.Sprintf("command timed out after %f seconds", t.Seconds())
}

func (t timeoutError) Timeout() time.Duration {
	return t.Duration
}

// ExitError is returned when a command sucessfully
// forked/executed but the process exits with a non-zero status
type ExitError interface {
	// ExitStatus best effort process exit code
	ExitStatus() int
}

type commandError struct {
	wrapped error
	status  int
}

func (e commandError) Error() string {
	return fmt.Sprintf("command exited with status %d: %s", e.status, e.wrapped)
}

func (e commandError) ExitStatus() int {
	return e.status
}

// Execute runs an ExecutionRequest
//
// Returns an error if the command does not run and exit with status 0
//
// Attempts to honor timeouts via context cancellation without abandoning
// child processes, but the world is messy and sensu intends to move on.
func (execution ExecutionRequest) Execute(ctx context.Context) error {
	if len(execution.Command) == 0 {
		return errEmptyCommand
	}
	executable, args := execution.Command[0], execution.Command[1:]
	var cmd *exec.Cmd
	cmd = exec.CommandContext(ctx, executable, args...)
	if len(execution.Env) > 0 {
		cmd.Env = execution.Env
	}
	if execution.Stdin != nil {
		cmd.Stdin = execution.Stdin
	}
	if execution.Stdout != nil {
		cmd.Stdout = execution.Stdout
	}
	if execution.Stderr != nil {
		cmd.Stderr = execution.Stderr
	}
	command.SetProcessGroup(cmd)
	start := time.Now()
	if err := cmd.Start(); err != nil {
		// Something unexpected happened when attempting to
		// fork/exec, return immediately.
		return err
	}
	// process was started.
	waitCh := make(chan struct{})
	var err error
	go func() {
		err = cmd.Wait()
		close(waitCh)
	}()

	// Wait for the process to complete or the context to be cancelled, whichever comes first.
	select {
	case <-waitCh:
		if err != nil {
			// The command most likely returned a non-zero exit status.
			if exitError, ok := err.(*exec.ExitError); ok {
				// Best effort to determine the exit status, this
				// should work on Linux, OSX, and Windows.
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					return &commandError{wrapped: err, status: status.ExitStatus()}
				}
				return &commandError{wrapped: err, status: command.FallbackExitStatus}
			}
			return &commandError{wrapped: err, status: command.FallbackExitStatus}
		}
		// process finished without error
		return nil
	case <-ctx.Done():
		if err := command.KillProcess(cmd); err != nil {
			killErr := fmt.Errorf("unable to kill process id #%d after command timeout: %s", cmd.Process.Pid, err.Error())
			logger := logrus.WithFields(logrus.Fields{"component": "commandv2"})
			logger.WithError(killErr).Errorf("Execution timed out - Unable to TERM/KILL the process: #%d. Potential zombie process.", cmd.Process.Pid)
			return timeoutError{Duration: time.Since(start), Wrapped: killErr}
		}
		return timeoutError{Duration: time.Since(start)}
	}
}

// ShellCommand builds a host appropriate shell command
// for use with ExecutionRequest.Command
func ShellCommand(command string) []string {
	return append(unixShellCommand, command)
}
