//go:build !windows
// +build !windows

// Package provides a unix-only system command execution
// and process limiting. Differs from sensu-go/command in
// that it does not assume a shell and has multiple options
// for handling timeouts appropriate for commands that may
// write to disk.
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
	// Timeout
	Timeout TimeoutStrategy
}

type TimeoutStrategy interface {
	// Signals when the timeout has been reached
	Signal() <-chan struct{}
	// Cleanup handles cleaning up any child processes
	Cleanup(ctx context.Context, cmd *exec.Cmd, waitErr <-chan error) error
}

// TimeoutError is returned when a command
// is interrupted by its timeout and the
// exit error could not be resolved
type TimeoutError interface {
	Timeout() time.Duration
}

type timeoutError struct {
	Duration time.Duration
	Err      error
}

func (t timeoutError) Error() string {
	e := fmt.Sprintf("command timed out after %f seconds", t.Duration.Seconds())
	if t.Err != nil {
		e = fmt.Sprintf("%s: %v", e, t.Err)
	}
	return e
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

type exitError struct {
	Err    error
	Status int
}

func (e exitError) Error() string {
	return fmt.Sprintf("command exited with status %d: %s", e.Status, e.Err)
}

func (e exitError) ExitStatus() int {
	return e.Status
}

// Execute runs an ExecutionRequest
//
// Returns an error if the command does not run and exit with status 0.
// The error will implement `ExitError` if the command was sucesfully started and
// waited on. Executions terminated by a timeout where an exit status could not be
// resolved will return errors implementing `TimeoutError`.
//
// Other error types are possible.
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
	if execution.Timeout == nil {
		execution.Timeout = TimeoutKillOnContextDone(ctx)
	}
	command.SetProcessGroup(cmd)
	start := time.Now()
	if err := cmd.Start(); err != nil {
		// Something unexpected happened when attempting to
		// fork/exec, return immediately.
		return err
	}
	// process was started.
	waitErrCh := make(chan error, 1)
	go func() {
		waitErrCh <- cmd.Wait()
		close(waitErrCh)
	}()

	// Wait for the process to complete or the context to be cancelled, whichever comes first.
	select {
	case waitErr := <-waitErrCh:
		return handleWaitErr(waitErr)
	case <-execution.Timeout.Signal():
		if err := execution.Timeout.Cleanup(ctx, cmd, waitErrCh); err != nil {
			if exErr, ok := err.(exitError); ok {
				// cleanup was able to await process and get exit error
				return exitError{
					Status: exErr.ExitStatus(),
					Err:    fmt.Errorf("timeout triggered process cleanup: %v", exErr.Err),
				}
			}
			killErr := fmt.Errorf("unable to clean up process id #%d after command timeout: %v", cmd.Process.Pid, err)
			return timeoutError{Duration: time.Since(start), Err: killErr}
		}
		return timeoutError{Duration: time.Since(start)}
	}
}

// ShellCommand builds a host appropriate shell command
// for use with ExecutionRequest.Command
func ShellCommand(command string) []string {
	return append(unixShellCommand, command)
}

func handleWaitErr(waitErr error) error {
	if waitErr != nil {
		// The command most likely returned a non-zero exit status.
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			// Best effort to determine the exit status, this
			// should work on Linux, OSX, and Windows.
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				return &exitError{Err: waitErr, Status: status.ExitStatus()}
			}
			return &exitError{Err: waitErr, Status: command.FallbackExitStatus}
		}
		return &exitError{Err: waitErr, Status: command.FallbackExitStatus}
	}
	// process finished without error
	return nil
}
