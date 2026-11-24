//go:build !windows

package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestExecuteDoesNotHangOnBackgroundChild(t *testing.T) {
	old := waitDelay
	waitDelay = 500 * time.Millisecond
	defer func() { waitDelay = old }()

	done := make(chan *ExecutionResponse, 1)
	go func() {
		resp, _ := (&ExecutionRequest{}).Execute(context.Background(), ExecutionRequest{Command: "sleep 30 & echo started"})
		done <- resp
	}()
	select {
	case resp := <-done:
		if resp.Status != 0 {
			t.Errorf("status = %d, want 0 (output %q)", resp.Status, resp.Output)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Execute blocked on a background child holding the output pipes")
	}
}

func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// A check with a Timeout runs in its own process group. When the shell exits but
// a background child keeps the output pipes open, Wait gives up after WaitDelay;
// the child must not be left running, one per check run.
func TestExecuteKillsBackgroundChildOnWaitDelay(t *testing.T) {
	old := waitDelay
	waitDelay = 300 * time.Millisecond
	defer func() { waitDelay = old }()

	pidFile := filepath.Join(t.TempDir(), "child.pid")
	req := ExecutionRequest{
		Command: fmt.Sprintf("sleep 60 & echo $! > %s; echo started", pidFile),
		Timeout: 30,
	}
	resp, err := (&ExecutionRequest{}).Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != 0 {
		t.Errorf("status = %d, want 0 (output %q)", resp.Status, resp.Output)
	}

	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = syscall.Kill(pid, syscall.SIGKILL) }()

	deadline := time.Now().Add(3 * time.Second)
	for processAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if processAlive(pid) {
		t.Errorf("background child %d is still running after Execute returned", pid)
	}
}

// Wait returns the ExitError in preference to ErrWaitDelay, so a check that exits
// non-zero while a background child holds the output pipes must still have that
// child cleaned up.
func TestExecuteKillsBackgroundChildWhenCheckExitsNonZero(t *testing.T) {
	old := waitDelay
	waitDelay = 300 * time.Millisecond
	defer func() { waitDelay = old }()

	pidFile := filepath.Join(t.TempDir(), "child.pid")
	req := ExecutionRequest{
		Command: fmt.Sprintf("sleep 60 & echo $! > %s; echo started; exit 2", pidFile),
		Timeout: 30,
	}
	resp, err := (&ExecutionRequest{}).Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != 2 {
		t.Errorf("status = %d, want 2 (output %q)", resp.Status, resp.Output)
	}

	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = syscall.Kill(pid, syscall.SIGKILL) }()

	deadline := time.Now().Add(3 * time.Second)
	for processAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if processAlive(pid) {
		t.Errorf("background child %d is still running after a non-zero exit", pid)
	}
}

func childPID(t *testing.T, pidFile string) int {
	t.Helper()
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

func waitGone(pid int) bool {
	deadline := time.Now().Add(2 * time.Second)
	for processAlive(pid) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	return !processAlive(pid)
}

// A handler that detaches its background process from the output pipes
// (nohup notifier >/dev/null 2>&1 &) has not leaked anything the command's
// pipes hold open, so it must keep running after Execute returns.
func TestExecuteLeavesDetachedBackgroundChildAlone(t *testing.T) {
	old := waitDelay
	waitDelay = 300 * time.Millisecond
	defer func() { waitDelay = old }()

	for _, tc := range []struct {
		name   string
		suffix string
		status int
	}{{"exit 0", "", 0}, {"exit 2", "; exit 2", 2}} {
		t.Run(tc.name, func(t *testing.T) {
			pidFile := filepath.Join(t.TempDir(), "child.pid")
			req := ExecutionRequest{
				Command: fmt.Sprintf("sleep 60 >/dev/null 2>&1 & echo $! > %s%s", pidFile, tc.suffix),
				Timeout: 30,
			}
			resp, err := (&ExecutionRequest{}).Execute(context.Background(), req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.Status != tc.status {
				t.Errorf("status = %d, want %d", resp.Status, tc.status)
			}
			pid := childPID(t, pidFile)
			defer func() { _ = syscall.Kill(pid, syscall.SIGKILL) }()
			time.Sleep(200 * time.Millisecond)
			if !processAlive(pid) {
				t.Errorf("detached background process %d was killed although it holds no output pipe", pid)
			}
		})
	}
}

// A command that simply runs for longer than WaitDelay and then exits non-zero
// is not a command that left pipes held open, so the group kill must not be
// taken from the elapsed time alone.
func TestExecuteLongRunningNonZeroCommandDoesNotKillDetachedChild(t *testing.T) {
	old := waitDelay
	waitDelay = 300 * time.Millisecond
	defer func() { waitDelay = old }()

	pidFile := filepath.Join(t.TempDir(), "child.pid")
	req := ExecutionRequest{
		Command: fmt.Sprintf("sleep 60 >/dev/null 2>&1 & echo $! > %s; sleep 1; exit 2", pidFile),
		Timeout: 30,
	}
	resp, err := (&ExecutionRequest{}).Execute(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != 2 {
		t.Errorf("status = %d, want 2", resp.Status)
	}
	pid := childPID(t, pidFile)
	defer func() { _ = syscall.Kill(pid, syscall.SIGKILL) }()
	time.Sleep(200 * time.Millisecond)
	if !processAlive(pid) {
		t.Errorf("detached background process %d was killed by a command that merely ran longer than WaitDelay", pid)
	}
}
