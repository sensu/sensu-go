package mockqueue

import (
	"context"

	"github.com/sensu/sensu-go/backend/queue"
	"github.com/stretchr/testify/mock"
)

// MockQueue ...
type MockQueue struct {
	mock.Mock
}

// Enqueue ...
func (m *MockQueue) Enqueue(ctx context.Context, value string) error {
	args := m.Called(ctx, value)
	return args.Error(0)
}

// Dequeue ...
func (m *MockQueue) Dequeue(ctx context.Context) (*queue.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).(*queue.Item), args.Error(1)
}
