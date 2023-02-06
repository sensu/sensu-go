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
func (m *MockQueue) Enqueue(ctx context.Context, value queue.Item) error {
	args := m.Called(ctx, value)
	return args.Error(0)
}

// Reserve ...
func (m *MockQueue) Reserve(ctx context.Context, name string) (queue.Reservation, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(queue.Reservation), args.Error(1)
}

type MockReservation struct {
	mock.Mock
}

func (m *MockReservation) Item() queue.Item {
	return m.Called().Get(0).(queue.Item)
}

func (m *MockReservation) Ack(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func (m *MockReservation) Nack(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
