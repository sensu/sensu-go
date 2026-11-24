// Package command provides system command execution.
package command

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sensu/sensu-go/types"
	bytesutil "github.com/sensu/sensu-go/util/bytes"
	"github.com/sirupsen/logrus"
)

const undocumentedTestCheckCommand = "!sensu_test_check!"

// waitDelay is how long Execute waits for the output pipes to close after the
// command's process has exited.
var waitDelay = 5 * time.Second

const cannedResponseText = `
                         .'loo:,
                        ,KNMMWNWX
                  ..    ,000OkxkW'
                 ,o,.   .O0KOOOk0:
                 dkl.    :OO0kkkk;
                 :ko     .lOOOOkx
              'OXX0:       'xWMdkd;o,.
              cMMMM; .,lkXN;oWNOk,;MMMMWXx:
              oMMMM:KNWMMMMl'.cl .NMMMMMMMMX
              NMMMWkMMMMMMMMWxxKONMMMMMMMMMM
             oMMMMMMMMMMMMMMMNW0NMMMMMMMMMMM.
             KMMMMMMMMMMMMMMWMWWMMMMMMMMMMMMN
             oKXXKKKKXMMMMMMMMMWMMMMMMMMMMMMM.
                     'MMMMMMMMMMMMMMMMMMMMMMMk
                     .MMMMMMMMWMMMMMMMMMMMMMMM
                      WMMMMMMMMMMWNMMMMMMMMMMN
                      WMMMMMMMMMWX0kO0WMMMMMMO
                     .MMMMMMMMMMMNX0kkWMMMMWO'
                     ;MMMMMMMMMMMMWXNNMMMMW.
`

var CannedResponse = &ExecutionResponse{
	Output: cannedResponseText,
}

// Executor ...
type Executor interface {
	Execute(context.Context, ExecutionRequest) (*ExecutionResponse, error)
}

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

// ExecutionRequest provides information about a system command execution,
// somewhat of an abstraction intended to be used for Sensu check,
// mutator, and handler execution.
type ExecutionRequest struct {
	// Command is the command to be executed.
	Command string

	// Env ...
	Env []string

	// Input to provide the command via STDIN.
	Input string

	// Execution timeout in seconds, will be set to a default if
	// not specified.
	Timeout int

	// Name is the name of the resource that is invoking the execution.
	Name string

	// InProgress is a map of checks that are still in execution, this is
	// necessary for a check or hook to escape zombie processes.
	InProgress map[string]*types.CheckConfig

	// InProgressMu is the mutex for the InProgress map.
	InProgressMu *sync.Mutex
}

// ExecutionResponse provides the response information of an ExecutionRequest.
type ExecutionResponse struct {
	// Combined command execution STDOUT/ERR.
	Output string

	// Command execution exit status.
	Status int

	// Duration provides command execution time in seconds.
	Duration float64
}

// NewExecutor ...
func NewExecutor() Executor {
	return &ExecutionRequest{}
}

// Execute executes a system command (fork/exec) with a
// timeout, optionally writing to STDIN, capturing its combined output
// (STDOUT/ERR) and exit status.
func (e *ExecutionRequest) Execute(ctx context.Context, execution ExecutionRequest) (*ExecutionResponse, error) {
	if execution.Command == undocumentedTestCheckCommand {
		return CannedResponse, nil
	}
	resp := &ExecutionResponse{}
	logger := logrus.WithFields(logrus.Fields{"component": "command"})
	// Using a platform specific shell to "cheat", as the shell
	// will handle certain failures for us, where golang exec is
	// known to have troubles, e.g. command not found. We still
	// use a fallback exit status in the unlikely event that the
	// exit status cannot be determined.
	var cmd *exec.Cmd

	// Use context.WithCancel for command execution timeout.
	// context.WithTimeout will not kill child/grandchild processes
	// (see issues tagged in https://github.com/sensu/sensu-go/issues/781).
	// Rather, we will use a timer, CancelFunc and proc functions
	// to perform full cleanup.
	ctx, timeout := context.WithCancel(ctx)
	defer timeout()

	// Taken from Sensu-Spawn (Sensu 1.x.x).
	cmd = Command(ctx, execution.Command)

	// Set the ENV for the command if it is set
	if len(execution.Env) > 0 {
		cmd.Env = execution.Env
	}

	// Share an output buffer between STDOUT/ERR, following the
	// Nagios plugin spec.
	var output bytesutil.SyncBuffer

	// The output pipe is created here instead of letting exec.Cmd do it, so that
	// we know for certain whether something the command left running still holds
	// it open after the command exits. Cmd.Wait cannot say: it reports that as
	// ErrWaitDelay only if the command exited successfully, and not at all if it
	// exited non-zero.
	outR, outW, pipeErr := os.Pipe()
	if pipeErr != nil {
		return resp, pipeErr
	}
	defer outR.Close()
	cmd.Stdout = outW
	cmd.Stderr = outW

	// If Input is specified, write to STDIN.
	if execution.Input != "" {
		cmd.Stdin = strings.NewReader(execution.Input)
	}

	started := time.Now()
	defer func() {
		resp.Duration = time.Since(started).Seconds()
	}()

	timer := time.NewTimer(math.MaxInt64)
	defer timer.Stop()
	if execution.Timeout != 0 {
		SetProcessGroup(cmd)
		timer.Stop()
		timer = time.NewTimer(time.Duration(execution.Timeout) * time.Second)
	}
	if err := cmd.Start(); err != nil {
		// Something unexpected happened when attempting to
		// fork/exec, return immediately.
		_ = outW.Close()
		return resp, err
	}
	// From here on only the command and what it starts hold the write end, so the
	// read end reaches EOF exactly when none of them are left.
	_ = outW.Close()

	// Read once: the goroutine below can outlive this call (a timeout returns
	// before Wait does), and waitDelay is a variable that tests change.
	delay := waitDelay

	copyDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, outR)
		close(copyDone)
	}()

	waitCh := make(chan struct{})
	var err error
	pipesHeldOpen := false
	go func() {
		err = cmd.Wait()
		select {
		case <-copyDone:
		case <-time.After(delay):
			// The command is gone, but something it left running still holds the
			// output pipe. Stop waiting for it.
			pipesHeldOpen = true
			_ = outR.Close()
			<-copyDone
		}
		close(waitCh)
	}()

	// Wait for the process to complete or the timer to trigger, whichever comes first.
	var killErr error
	select {
	case <-waitCh:
		// Only what is still holding the command's output pipe is cleaned up, so
		// that one process per run is not leaked. Anything that detached from the
		// pipes (nohup cmd >/dev/null 2>&1 &) is left running. Without a timeout
		// the command has no process group of its own to signal.
		if execution.Timeout != 0 && pipesHeldOpen {
			if killErr := KillProcess(cmd); killErr != nil && !errors.Is(killErr, syscall.ESRCH) {
				logger.WithError(killErr).Warn("unable to kill the processes left behind by the command")
			}
		}
		resp.Output = output.String()
		if err != nil {
			// The command most likely return a non-zero exit status.
			if exitError, ok := err.(*exec.ExitError); ok {
				// Best effort to determine the exit status, this
				// should work on Linux, OSX, and Windows.
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					resp.Status = status.ExitStatus()
				} else {
					resp.Status = FallbackExitStatus
				}
			} else {
				resp.Status = FallbackExitStatus
			}
		} else {
			// Everything is A-OK.
			resp.Status = OKExitStatus
		}

	case <-timer.C:
		var killErrOutput string
		if killErr = KillProcess(cmd); killErr != nil {
			logger.WithError(killErr).Errorf("Execution timed out - Unable to TERM/KILL the process: #%d", cmd.Process.Pid)
			killErrOutput = fmt.Sprintf("Unable to TERM/KILL the process: #%d\n", cmd.Process.Pid)
			escapeZombie(&execution)
		}
		timeout()
		resp.Output = fmt.Sprintf("%s%s%s", TimeoutOutput, killErrOutput, output.String())
		resp.Status = TimeoutExitStatus
	}

	return resp, nil
}

func escapeZombie(ex *ExecutionRequest) {
	logger := logrus.WithFields(logrus.Fields{"component": "command"})
	if ex.InProgress != nil && ex.InProgressMu != nil && ex.Name != "" {
		logger.WithField("check", ex.Name).Warn("check or hook execution created zombie process - escaping in order for the check to execute again")
		ex.InProgressMu.Lock()
		delete(ex.InProgress, ex.Name)
		ex.InProgressMu.Unlock()
	} else {
		logger.Error("unable to escape zombie process created from command execution")
	}
}

// FixtureExecutionResponse returns an Execution for use in testing
func FixtureExecutionResponse(status int, output string) *ExecutionResponse {
	return &ExecutionResponse{
		Output:   output,
		Status:   status,
		Duration: 1,
	}
}
