package exec

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var cmdPrefix []string = []string{"cmd", "/c"}

// ShellCommand builds a host appropriate shell command
// for use with ExecutionRequest.Command
func ShellCommand(command string) []string {
	return append(cmdPrefix, command)
}

// SetProcessGroup sets the process group of the command process
func SetProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr.CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP
}

// SignalTerminate signal the command process and any children to
// shut down
func SignalTerminate(cmd *exec.Cmd) error {
	process := cmd.Process
	if process == nil {
		return nil
	}

	err := Command(context.Background(), fmt.Sprintf("taskkill /T /PID %d", process.Pid)).Run()
	if err == nil {
		return nil
	}
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
