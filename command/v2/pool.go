package v2

import (
	"context"
	"errors"

	"golang.org/x/sync/semaphore"
)

var (
	ErrExecutionPoolFull = errors.New("execution pool is full")
)

type ExecutionPool struct {
	sem   *semaphore.Weighted
	block bool
}

func NewExecutionPool(capacity int64, blockForExecution bool) *ExecutionPool {
	return &ExecutionPool{
		sem:   semaphore.NewWeighted(capacity),
		block: blockForExecution,
	}
}

// Execute reserves a spot in the pool and
// executes the ExecutionRequest
//
// Returns error when unable to reserve a spot in the pool.
// Otherwise behaves like ExecutionRequest.Execute
func (e *ExecutionPool) Execute(ctx context.Context, execution ExecutionRequest) error {
	if e.block {
		if err := e.sem.Acquire(ctx, 1); err != nil {
			return err
		}
	} else {
		if ok := e.sem.TryAcquire(1); !ok {
			return ErrExecutionPoolFull
		}
	}
	defer e.sem.Release(1)
	return execution.Execute(ctx)
}
