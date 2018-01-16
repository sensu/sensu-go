// Package command provides system command execution.
package command

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const (
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

	// Env ...
	Env []string

	// Input to provide the command via STDIN.
	Input string

	// Execution timeout in seconds, will be set to a default if
	// not specified.
	Timeout int

	// Combined command execution STDOUT/ERR.
	Output string

	// Command execution exit status.
	Status int

	// Duration provides command execution time in seconds.
	Duration float64
}

// ExecuteCommand executes a system command (fork/exec) with a
// timeout, optionally writing to STDIN, capturing its combined output
// (STDOUT/ERR) and exit status.
func ExecuteCommand(ctx context.Context, execution *Execution) (*Execution, error) {
	// Using a platform specific shell to "cheat", as the shell
	// will handle certain failures for us, where golang exec is
	// known to have troubles, e.g. command not found. We still
	// use a fallback exit status in the unlikely event that the
	// exit status cannot be determined.
	var cmd *exec.Cmd

	// Use the context deadline for command execution timeout.
	// This will be effectively ignored if the context already has
	// an earlier deadline, which is super rad.
	var timeout context.CancelFunc
	if execution.Timeout == 0 {
		ctx, timeout = context.WithCancel(ctx)
	} else {
		ctx, timeout = context.WithTimeout(ctx, time.Duration(execution.Timeout)*time.Second)
	}
	defer timeout()

	// Taken from Sensu-Spawn (Sensu 1.x.x).
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", execution.Command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", execution.Command)
	}

	// Set the ENV for the command if it is set
	if len(execution.Env) > 0 {
		cmd.Env = execution.Env
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

	started := time.Now()
	defer func() {
		execution.Duration = time.Since(started).Seconds()
	}()

	// Kill process and all of its children when the timeout has expired.
	// context.WithTimeout will not kill child/grandchild processes
	// (see issues tagged in https://github.com/sensu/sensu-go/issues/781),
	// rather we will use a timer and utility_os package to perform full cleanup.
	if execution.Timeout != 0 {
		SetProcessGroup(cmd)
		time.AfterFunc(time.Duration(execution.Timeout)*time.Second, func() {
			if err := KillProcess(cmd); err != nil {
				return
			}
		})
	}

	if err := cmd.Start(); err != nil {
		// Something unexpected happended when attepting to
		// fork/exec, return immediately.
		return execution, err
	}

	err := cmd.Wait()

	execution.Output = output.String()

	// The command execution timed out if the context deadline was
	// exceeded.
	if ctx.Err() == context.DeadlineExceeded {
		execution.Output = TimeoutOutput
		execution.Status = TimeoutExitStatus
	} else if err != nil {
		// The command most likely return a non-zero exit status.
		if exitError, ok := err.(*exec.ExitError); ok {
			// Best effort to determine the exit status, this
			// should work on Linux, OSX, and Windows.
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
		execution.Status = 0
	}

	return execution, nil
}
