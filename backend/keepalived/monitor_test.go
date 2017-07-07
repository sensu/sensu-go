package keepalived

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCreator struct {
	mock.Mock
}

func (m *mockCreator) Warn(e *types.Entity) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *mockCreator) Critical(e *types.Entity) error {
	args := m.Called(e)
	return args.Error(0)
}
func (m *mockCreator) Resolve(e *types.Entity) error {
	args := m.Called(e)
	return args.Error(0)
}

type mockDeregisterer struct {
	mock.Mock
}

func (m *mockDeregisterer) Deregister(e *types.Entity) error {
	args := m.Called(e)
	return args.Error(0)
}

func TestMonitorUpdate(t *testing.T) {
	assert := assert.New(t)

	mockStore := &mockstore.MockStore{}

	entity := types.FixtureEntity("entity")
	event := &types.Event{
		Entity: entity,
	}

	monitor := &KeepaliveMonitor{
		Entity: entity,
		Store:  mockStore,
	}
	monitor.Start()

	mockStore.On("UpdateEntity", mock.Anything, entity).Return(nil)
	mockStore.On("UpdateKeepalive", mock.Anything, entity.ID, mock.AnythingOfType("int64")).Return(nil)

	assert.NoError(monitor.Update(event))
}

func TestStop(t *testing.T) {
	monitor := &KeepaliveMonitor{
		reset: make(chan interface{}),
	}
	monitor.Stop()
	assert.True(t, monitor.IsStopped(), "IsStopped returns true if stopped")
}

func TestMonitorDeregistration(t *testing.T) {
	entity := types.FixtureEntity("entity")
	entity.KeepaliveTimeout = 0
	entity.Deregister = true
	dereg := &mockDeregisterer{}
	dereg.On("Deregister", entity).Return(nil)

	monitor := &KeepaliveMonitor{
		Entity:       entity,
		Deregisterer: dereg,
	}

	monitor.Start()
	time.Sleep(100 * time.Millisecond)
	dereg.AssertCalled(t, "Deregister", entity)
	assert.True(t, monitor.IsStopped(), "monitor is stopped after deregistration")
}

func TestMonitorAlert(t *testing.T) {
	entity := types.FixtureEntity("entity")
	entity.KeepaliveTimeout = 0
	entity.Deregister = false
	creator := &mockCreator{}
	creator.On("Warn", entity).Return(nil)

	monitor := &KeepaliveMonitor{
		Entity:       entity,
		EventCreator: creator,
	}

	monitor.Start()
	time.Sleep(100 * time.Millisecond)
	creator.AssertCalled(t, "Warn", entity)
}
