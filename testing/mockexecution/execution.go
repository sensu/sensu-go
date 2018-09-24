package mockexecution

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/command"
)

// True mocks a command execution that returns exit status 0
func True(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	execution.Status = 0
	return execution, nil
}

// False mocks a command execution that returns exit status 1
func False(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	execution.Status = 1
	return execution, nil
}

// Timeout mocks a command execution that returns exit status 2
func Timeout(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	execution.Status = 2
	return execution, nil
}

// Error mocks a command execution that returns exit status 127 and execution error
func Error(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	execution.Status = 127
	return execution, errors.New("command not found")
}

// Metrics mocks a command execution that outputs metrics
func Metrics(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	execution.Status = 0
	execution.Output = "metric.foo 1 123456789\nmetric.bar 2 987654321"
	return execution, nil
}

// Hello mocks a command execution that prints 'hello'
func Hello(ctx context.Context, execution *command.Execution) (*command.Execution, error) {
	execution.Status = 0
	execution.Output = "hello"
	return execution, nil
}
