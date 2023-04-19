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
func (m *MockQueue) Dequeue(ctx context.Context) (queue.QueueItem, error) {
	args := m.Called(ctx)
	return args.Get(0).(queue.QueueItem), args.Error(1)
}

// Getter ...
type Getter struct {
	mock.Mock
}

// GetQueue ...
func (g *Getter) GetQueue(path ...string) queue.Interface {
	ifaceArgs := make([]interface{}, len(path))
	for i := range path {
		ifaceArgs[i] = path[i]
	}
	args := g.Called(ifaceArgs...)
	return args.Get(0).(queue.Interface)
}
