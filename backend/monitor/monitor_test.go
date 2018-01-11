// +build !integration

package monitor

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type testHandler struct{}

// create update and failure handlers for use with the monitor
func (handler *testHandler) HandleUpdate(event *types.Event) error {
	return nil
}

func (handler *testHandler) HandleFailure(event *types.Entity) error {
	return nil
}

func TestMonitorUpdate(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("entity")
	event := &types.Event{
		Entity: entity,
	}

	handler := &testHandler{}

	monitor := &Monitor{
		Entity:         entity,
		Timeout:        60 * time.Second,
		FailureHandler: handler,
		UpdateHandler:  handler,
	}

	monitor.Start()

	assert.NoError(monitor.HandleUpdate(event))
}

func TestMonitorStop(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("entity")

	handler := &testHandler{}
	monitor := &Monitor{Entity: entity,
		Timeout:        60 * time.Second,
		FailureHandler: handler,
		UpdateHandler:  handler,
	}
	monitor.reset = make(chan interface{})
	monitor.Stop()
	assert.True(monitor.IsStopped(), "IsStopped returns true if stopped")
}

func TestMonitorReset(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("entity")
	entity.KeepaliveTimeout = 0
	entity.Deregister = false

	event := types.FixtureEvent("entity", "keepalive")
	event.Entity.KeepaliveTimeout = 120

	handler := &testHandler{}

	monitor := &Monitor{Entity: entity,
		Timeout:        60 * time.Second,
		FailureHandler: handler,
		UpdateHandler:  handler,
	}

	monitor.Reset(time.Now().Unix())
	time.Sleep(100 * time.Millisecond)
	assert.False(monitor.IsStopped())
}
