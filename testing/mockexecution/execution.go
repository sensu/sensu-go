package mockexecution

import (
	"context"

	"github.com/sensu/sensu-go/command"
	"github.com/stretchr/testify/mock"
)

// MockExecution ...
type MockExecution struct {
	mock.Mock
}

// Execute ...
func (e *MockExecution) Execute(ctx context.Context, execution *command.ExecutionRequest) (*command.ExecutionResponse, error) {
	args := e.Called(ctx, execution)
	return args.Get(0).(*command.ExecutionResponse), args.Error(1)
}
