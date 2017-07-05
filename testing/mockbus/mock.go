package mockbus

import "github.com/stretchr/testify/mock"

// MockBus ...
type MockBus struct {
	mock.Mock
}

// Start ...
func (m *MockBus) Start() error {
	args := m.Called()
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

// Subscribe ...
func (m *MockBus) Subscribe(topic string, consumer string, channel chan<- interface{}) error {
	args := m.Called(topic, consumer, channel)
	return args.Error(0)
}

// Publish ...
func (m *MockBus) Publish(topic string, message interface{}) error {
	args := m.Called(topic, message)
	return args.Error(0)
}

// Unsubscribe ...
func (m *MockBus) Unsubscribe(topic, consumer string) error {
	args := m.Called(topic, consumer)
	return args.Error(0)
}
