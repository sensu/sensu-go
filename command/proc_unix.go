// +build !windows

package command

import (
	"os/exec"
	"syscall"
)

// SetProcessGroup sets the process group of the command process
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// KillProcess kills the command process and any child processes
func KillProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
