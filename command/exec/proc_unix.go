//go:build !windows
// +build !windows

package exec

import (
	"os/exec"
	"syscall"
)

var unixShellCommand []string = []string{"sh", "-c"}

// SignalTerminate signal the command process and any children to
// shut down
func SignalTerminate(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
}

// KillProcess kills the command process and any child processes
func KillProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}

// ShellCommand builds a host appropriate shell command
// for use with ExecutionRequest.Command
func ShellCommand(command string) []string {
	return append(unixShellCommand, command)
}

// SetProcessGroup sets the process group of the command process
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGTERM}
}
