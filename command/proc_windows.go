// +build windows

package command

import (
	"os/exec"
)

// SetProcessGroup sets the process group of the command process
func SetProcessGroup(cmd *exec.Cmd) {}

// KillProcess kills the command process and any child processes
func KillProcess(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
