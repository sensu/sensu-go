package mockmonitor

import (
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

// MockMonitor ...
type MockMonitor struct {
	mock.Mock
}

// Stop ...
func (m *MockMonitor) Stop() {
}

// IsStopped ...
func (m *MockMonitor) IsStopped() bool {
	args := m.Called()
	return args.Bool(0)
}

// HandleUpdate ...
func (m *MockMonitor) HandleUpdate(e *types.Event) error {
	args := m.Called(e)
	return args.Error(0)
}

// HandleFailure ...
func (m *MockMonitor) HandleFailure(entity *types.Entity, event *types.Event) error {
	args := m.Called(entity, event)
	return args.Error(0)
}

// GetTimeout ...
func (m *MockMonitor) GetTimeout() time.Duration {
	return time.Duration(120 * time.Second)
}
