// Package command provides system command execution.
package command

import (
	"bytes"
	"os/exec"
	"strings"
	"syscall"
)

const fallbackExitStatus int = 3

// Execution provides information about a system command execution,
// somewhat of an abstraction intended to be used for Sensu check,
// mutator, and handler execution.
type Execution struct {
	// Command is the command to be executed.
	Command string

	// Input to provide the command via STDIN.
	Input string

	// Combined command execution STDOUT/ERR.
	Output string

	// Command execution exit status.
	Status int
}

// ExecuteCommand executes a system command (fork/exec), optionally
// writing to STDIN, capture its combined output (STDOUT/ERR) and exit
// status.
func ExecuteCommand(c *Execution) (*Execution, error) {
	// Using a platform specific shell to "cheat", as the shell
	// will handle certain failures for us, where golang exec is
	// known to have troubles, e.g. command not found. We still
	// use a fallback exit status in the unlikely event that the
	// exit status cannot be determined.
	cmd := exec.Command("sh", "-c", c.Command)

	// Try to sharing an output buffer between STDOUT/ERR.
	var output bytes.Buffer

	cmd.Stdout = &output
	cmd.Stderr = &output

	if c.Input != "" {
		cmd.Stdin = strings.NewReader(c.Input)
	}

	if err := cmd.Start(); err != nil {
		// Something unexpected happended when attepting to
		// fork/exec, return immediately.
		return c, err
	}

	err := cmd.Wait()

	c.Output = output.String()

	// The command most likely return a non-zero exit status.
	if err != nil {
		// Best effort to determine the exit status, this
		// should work on Linux, OSX, and Windows.
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				c.Status = status.ExitStatus()
			} else {
				c.Status = fallbackExitStatus
			}
		} else {
			c.Status = fallbackExitStatus
		}
	} else {
		// Everything is A-OK.
		c.Status = 0
	}

	return c, nil
}
