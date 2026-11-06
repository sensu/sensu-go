//go:build !windows

package command

import (
	"context"
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
