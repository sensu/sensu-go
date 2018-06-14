package monitor

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type testMonitorsHandler struct{}

// create update and failure handlers for use with the monitor
func (handler *testMonitorsHandler) HandleUpdate(event *types.Event) error {
	if event.Entity.ID == "entity" {
		return nil
	}
	return errors.New("test update handler error")
}

func (handler *testMonitorsHandler) HandleFailure(entity *types.Entity, event *types.Event) error {
	if entity.ID == "entity" {
		return nil
	}
	return errors.New("test failure handler error")
}

func TestGetMonitors(t *testing.T) {

}

func TestMonitorsHandleUpdate(t *testing.T) {

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

			monitor := NewMonitors()

			assert.EqualValues(tc.expectedErr, monitor.HandleUpdate(tc.event))
		})
	}

}
