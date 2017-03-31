// Package command provides system command execution.
package command

import (
	"bytes"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
	// DefaultTimeout specifies the default command execution
	// timeout in seconds.
	DefaultTimeout int = 60

	// TimeoutOutput specifies the command execution output in the
	// event of an execution timeout.
	TimeoutOutput string = "Execution timed out\n"

	// OKExitStatus specifies the command execution exit status
	// that indicates a success, A-OK.
	OKExitStatus int = 0

	// TimeoutExitStatus specifies the command execution exit
	// status in the event of an execution timeout.
	TimeoutExitStatus int = 2

	// FallbackExitStatus specifies the command execution exit
	// status used when golang is unable to determine the exit
	// status.
	FallbackExitStatus int = 3
)

// Execution provides information about a system command execution,
// somewhat of an abstraction intended to be used for Sensu check,
// mutator, and handler execution.
type Execution struct {
	// Command is the command to be executed.
	Command string

	// Input to provide the command via STDIN.
	Input string

	// Execution timeout in seconds, will be set to a default if
	// not specified.
	Timeout int

	// Combined command execution STDOUT/ERR.
	Output string

	// Command execution exit status.
	Status int
}

// ExecuteCommand executes a system command (fork/exec), optionally
// writing to STDIN, capture its combined output (STDOUT/ERR) and exit
// status.
func ExecuteCommand(execution *Execution) (*Execution, error) {
	// Using a platform specific shell to "cheat", as the shell
	// will handle certain failures for us, where golang exec is
	// known to have troubles, e.g. command not found. We still
	// use a fallback exit status in the unlikely event that the
	// exit status cannot be determined.
	var cmd *exec.Cmd

	// Taken from Sensu-Spawn (Sensu 1.x.x).
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", execution.Command)
	} else {
		cmd = exec.Command("sh", "-c", execution.Command)
	}

	// Share an output buffer between STDOUT/ERR, following the
	// Nagios plugin spec.
	var output bytes.Buffer

	cmd.Stdout = &output
	cmd.Stderr = &output

	// If Input is specified, write to STDIN.
	if execution.Input != "" {
		cmd.Stdin = strings.NewReader(execution.Input)
	}

	if err := cmd.Start(); err != nil {
		// Something unexpected happended when attepting to
		// fork/exec, return immediately.
		return execution, err
	}

	// If Timeout is not specified, use the default.
	if execution.Timeout == 0 {
		execution.Timeout = DefaultTimeout
	}

	// Use a goroutine and channel for execution timeout.
	done := make(chan error, 1)
	defer close(done)
	go func() {
		// Wait for the command execution to complete.
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(time.Duration(execution.Timeout) * time.Second):
		// Attempt to kill -9 the command execution child process.
		if err := cmd.Process.Kill(); err != nil {
			return execution, err
		}

		execution.Output = TimeoutOutput
		execution.Status = TimeoutExitStatus
	case err := <-done:
		execution.Output = output.String()

		// The command most likely return a non-zero exit status.
		if err != nil {
			// Best effort to determine the exit status, this
			// should work on Linux, OSX, and Windows.
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					execution.Status = status.ExitStatus()
				} else {
					execution.Status = FallbackExitStatus
				}
			} else {
				execution.Status = FallbackExitStatus
			}
		} else {
			// Everything is A-OK.
			execution.Status = OKExitStatus
		}
	}

	return execution, nil
}
