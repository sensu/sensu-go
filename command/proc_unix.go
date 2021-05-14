// +build !windows

package command

import (
	"context"
	"os/exec"
	"syscall"
)

// Command returns a command to execute a script through a shell.
func Command(ctx context.Context, command string) *exec.Cmd {
	return exec.CommandContext(ctx, "sh", "-c", command)
}

// KillProcess kills the command process and any child processes
func KillProcess(cmd *exec.Cmd) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
