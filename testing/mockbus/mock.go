package mockbus

import (
	"context"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/stretchr/testify/mock"
)

// MockBus ...
type MockBus struct {
	mock.Mock
}

// Start ...
func (m *MockBus) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Stop ...
func (m *MockBus) Stop() error {
	args := m.Called()
	return args.Error(0)
}

// Status ...
func (m *MockBus) Status() error {
	args := m.Called()
	return args.Error(0)
}

// Err ...
func (m *MockBus) Err() <-chan error {
	args := m.Called()
	return args.Get(0).(<-chan error)
}

// Name ...
func (m *MockBus) Name() string {
	args := m.Called()
	return args.String(0)
}

// Subscribe ...
func (m *MockBus) Subscribe(topic string, consumer string, subscriber messaging.Subscriber) (messaging.Subscription, error) {
	args := m.Called(topic, consumer, subscriber)
	return args.Get(0).(messaging.Subscription), args.Error(1)
}

// Publish ...
func (m *MockBus) Publish(topic string, message interface{}) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

// PublishDirect ...
func (m *MockBus) PublishDirect(ctx context.Context, topic string, message interface{}) error {
	args := m.Called(ctx, topic, message)
	return args.Error(0)
}
