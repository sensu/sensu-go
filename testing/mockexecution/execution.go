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

// ExecuteCommand ...
func (e *MockExecution) ExecuteCommand(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	args := e.Called(ctx, execution)
	return args.Get(0).(*command.Execution), args.Error(1)
}
