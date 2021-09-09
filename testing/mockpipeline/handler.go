package mockpipeline

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

// HandlerAdapter ...
type HandlerAdapter struct {
	mock.Mock
}

// Name ...
func (m *HandlerAdapter) Name() string {
	args := m.Called()
	return args.Get(0).(string)
}

// CanHandle ...
func (m *HandlerAdapter) CanHandle(ref *corev2.ResourceReference) bool {
	args := m.Called(ref)
	return args.Get(0).(bool)
}

// Handle ...
func (m *HandlerAdapter) Handle(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event, data []byte) error {
	args := m.Called(ctx, ref, event, data)
	return args.Error(0)
}
