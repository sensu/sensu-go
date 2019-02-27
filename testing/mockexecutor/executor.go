package mockexecutor

import (
	"context"
	"sync"

	"github.com/sensu/sensu-go/command"
)

// MockExecutor ...
type MockExecutor struct {
	// Note: this used to use the testify mock.Mock type, but I removed that,
	// because it was causing race conditions to occur in tests, since the mock
	// library wanted to inspect fields that were being guarded by mutexes.
	response *command.ExecutionResponse
	err      error
	mu       sync.Mutex
}

func (e *MockExecutor) Return(r *command.ExecutionResponse, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.response = r
	e.err = err
}

// Execute ...
func (e *MockExecutor) Execute(ctx context.Context, execution command.ExecutionRequest) (*command.ExecutionResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.response, e.err
}
