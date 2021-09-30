package mockexecutor

import (
	"context"
	"sync"

	"github.com/sensu/sensu-go/command"
)

type RequestFunc func(context.Context, command.ExecutionRequest)

// MockExecutor ...
type MockExecutor struct {
	// Note: this used to use the testify mock.Mock type, but I removed that,
	// because it was causing race conditions to occur in tests, since the mock
	// library wanted to inspect fields that were being guarded by mutexes.
	requestFunc RequestFunc
	response    *command.ExecutionResponse
	err         error
	mu          sync.Mutex
}

func (e *MockExecutor) Return(r *command.ExecutionResponse, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.response = r
	e.err = err
}

func (e *MockExecutor) UnsafeReturn(r *command.ExecutionResponse, err error) {
	e.response = r
	e.err = err
}

func (e *MockExecutor) SetRequestFunc(fn RequestFunc) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.requestFunc = fn
}

// Execute ...
func (e *MockExecutor) Execute(ctx context.Context, execution command.ExecutionRequest) (*command.ExecutionResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.requestFunc != nil {
		e.requestFunc(ctx, execution)
	}
	return e.response, e.err
}
