//go:build windows
// +build windows

package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// Command returns a command to execute a script through a shell.
func Command(ctx context.Context, command string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "cmd")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// Manually set the command line arguments so they are not escaped
		// https://github.com/golang/go/commit/f18a4e9609aac3aa83d40920c12b9b45f9376aea
		// http://www.josephspurrier.com/prevent-escaping-exec-command-arguments-in-go/
		CmdLine: strings.Join([]string{"/c", command}, " "),
	}
	return cmd
}

// KillProcess kills the command process and any child processes
func KillProcess(cmd *exec.Cmd) error {
	process := cmd.Process
	if process == nil {
		return nil
	}

	err := Command(context.Background(), fmt.Sprintf("taskkill /T /F /PID %d", process.Pid)).Run()
	if err == nil {
		return nil
	}

	err = forceKill(process)
	if err == nil {
		return nil
	}
	err = process.Signal(os.Kill)

	return fmt.Errorf("could not kill process")
}

func forceKill(process *os.Process) error {
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, true, uint32(process.Pid))
	if err != nil {
		return err
	}

	err = syscall.TerminateProcess(handle, 0)
	_ = syscall.CloseHandle(handle)

	return err
}
