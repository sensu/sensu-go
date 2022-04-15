//go:build !windows
// +build !windows

package v2

import (
	"context"
	"testing"
	"time"

	"github.com/sensu/sensu-go/command"
)

func TestExecutorPoolConcurrencyLimits(t *testing.T) {
	testCtx := context.Background()

	type result struct {
		ExecutionResult *command.ExecutionResponse
		Error           error
	}

	// Attempts to start #ConcurrentExecutions Executions
	// of `sleep 10` which should all time out.
	// Asserts that #ExpectedExecutions fork/execs were ran
	// (and by inference that #ConcurrentExecutions - #ExpectedExecutions
	//  attempts were denied with ErrExecutionPoolFull)
	testCases := []struct {
		Name                 string
		BufferSize           int
		ConcurrentExecutions int
		ExpectedExecutions   int
	}{
		{
			Name:                 "serial execution",
			BufferSize:           1,
			ConcurrentExecutions: 1024,
			ExpectedExecutions:   1,
		},
		{
			Name:                 "excess bandwith",
			BufferSize:           100,
			ConcurrentExecutions: 10,
			ExpectedExecutions:   10,
		},
		{
			Name:                 "10x",
			BufferSize:           10,
			ConcurrentExecutions: 100,
			ExpectedExecutions:   10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			bufferSize, concurrentExecutions := tc.BufferSize, tc.ConcurrentExecutions
			p := NewExecutionPool(int64(bufferSize))

			results := make(chan result, concurrentExecutions)
			ctx, cancel := context.WithTimeout(testCtx, time.Millisecond*100)
			defer cancel()
			// Start N goroutines each trying to run a command that should time out
			for i := 0; i < concurrentExecutions; i++ {
				go func() {
					r, err := p.Execute(ctx, ExecutionRequest{
						Command: []string{"sleep", "10"},
					})
					results <- result{
						ExecutionResult: r,
						Error:           err,
					}
				}()
			}
			waitingCt := 0
			ranCt := 0
			for i := 0; i < concurrentExecutions; i++ {
				r := <-results
				if r.Error == ErrExecutionPoolFull {
					waitingCt++
				} else {
					ranCt++
				}
			}
			if ranCt != tc.ExpectedExecutions {
				t.Errorf("expected %d execution got %d", tc.ExpectedExecutions, ranCt)
			}

		})
	}
}

func TestExecutionPoolTimeout(t *testing.T) {
	pool := NewExecutionPool(1)

	timeout := time.Millisecond * 25
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	start := time.Now()
	defer cancel()
	r, err := pool.Execute(ctx, ExecutionRequest{
		Command: ShellCommand("sleep 1"),
	})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	if r.Status == 0 {
		t.Error("expected non-zero exit code")
	}
	elapsedTime := time.Since(start)
	if elapsedTime > (timeout+timeout/4) || (timeout-timeout/4) > elapsedTime {
		t.Errorf("expected timeout within 25%% of %d. was %d.", timeout, elapsedTime)
	}
}
