package v2

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/command"
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

func (e *ExecutionPool) Execute(ctx context.Context, execution ExecutionRequest) (*command.ExecutionResponse, error) {
	if e.block {
		if err := e.sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
	} else {
		if ok := e.sem.TryAcquire(1); !ok {
			return nil, ErrExecutionPoolFull
		}
	}
	defer e.sem.Release(1)
	return execute(ctx, execution)
}
