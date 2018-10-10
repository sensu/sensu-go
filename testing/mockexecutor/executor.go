package mockexecutor

import (
	"context"

	"github.com/sensu/sensu-go/command"
	"github.com/stretchr/testify/mock"
)

// MockExecutor ...
type MockExecutor struct {
	mock.Mock
}

// Execute ...
func (e *MockExecutor) Execute(ctx context.Context, execution command.ExecutionRequest) (*command.ExecutionResponse, error) {
	args := e.Called(ctx, execution)
	return args.Get(0).(*command.ExecutionResponse), args.Error(1)
}
