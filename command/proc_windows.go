// +build windows

package command

import (
	"context"
	"os/exec"
	"strings"
	"syscall"
)

// Command returns a command to execute a script through a shell.
func Command(ctx context.Context, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "cmd", "/c", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine: strings.Join(cmd.Args, " "),
	}
	return cmd
}

// SetProcessGroup sets the process group of the command process
func SetProcessGroup(cmd *exec.Cmd) {}

// KillProcess kills the command process and any child processes
func KillProcess(cmd *exec.Cmd) error {
	return cmd.Process.Kill()
}
