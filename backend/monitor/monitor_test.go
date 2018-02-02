package monitor

import (
	"errors"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type testHandler struct{}

// create update and failure handlers for use with the monitor
func (handler *testHandler) HandleUpdate(event *types.Event) error {
	if event.Entity.ID == "entity" {
		return nil
	}
	return errors.New("test update handler error")
}

func (handler *testHandler) HandleFailure(entity *types.Entity, event *types.Event) error {
	if entity.ID == "entity" {
		return nil
	}
	return errors.New("test failure handler error")
}

func TestMonitorHandleUpdate(t *testing.T) {

	passEntity := types.FixtureEntity("entity")
	failEntity := types.FixtureEntity("fail")
	testCases := []struct {
		name        string
		entity      *types.Entity
		event       *types.Event
		handler     *testHandler
		expectedErr error
	}{
		{
			name:   "test no error",
			entity: passEntity,
			event: &types.Event{
				Entity: passEntity,
			},
			handler:     &testHandler{},
			expectedErr: nil,
		},
		{
			name:   "test with error",
			entity: failEntity,
			event: &types.Event{
				Entity: failEntity,
			},
			handler:     &testHandler{},
			expectedErr: errors.New("test update handler error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			monitor := New(tc.entity, nil, (60 * time.Second), tc.handler, tc.handler)

			assert.EqualValues(tc.expectedErr, monitor.HandleUpdate(tc.event))
		})
	}

}

func TestMonitorHandleFailure(t *testing.T) {

	testCases := []struct {
		name        string
		entity      *types.Entity
		handler     *testHandler
		expectedErr error
	}{
		{
			name:        "test no error",
			entity:      types.FixtureEntity("entity"),
			handler:     &testHandler{},
			expectedErr: nil,
		},
		{
			name:        "test with error",
			entity:      types.FixtureEntity("fail"),
			handler:     &testHandler{},
			expectedErr: errors.New("test failure handler error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			monitor := New(tc.entity, nil, (60 * time.Second), tc.handler, tc.handler)

			assert.EqualValues(tc.expectedErr, monitor.HandleFailure(tc.entity, nil))
		})
	}

}

func TestMonitorStartStop(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("entity")

	handler := &testHandler{}
	monitor := New(entity, nil, (60 * time.Second), handler, handler)
	assert.Equal((60 * time.Second), monitor.Timeout)
	assert.False(monitor.IsStopped(), "IsStopped returns false if not stopped")
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

	monitor := New(entity, nil, (60 * time.Second), handler, handler)
	monitor.reset(0)
	time.Sleep(100 * time.Millisecond)
	assert.False(monitor.IsStopped())
}
