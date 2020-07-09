package agentd

import (
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/stretchr/testify/mock"
)

type mockWatcher struct {
	resultChan chan store.WatchEvent
}

func (w *mockWatcher) Result() <-chan store.WatchEvent {
	return w.resultChan
}

func (w *mockWatcher) Stop() {
	close(w.resultChan)
}

func Test_handleResults(t *testing.T) {
	type busFunc func(*mockbus.MockBus)

	// The state bytes are used to mock an invalid struct
	state, _ := wrap.Resource(corev3.FixtureEntityState("foo"))
	stateBytes, _ := proto.Marshal(state)

	cfg, _ := wrap.Resource(corev3.FixtureEntityConfig("bar"))
	cfgBytes, _ := proto.Marshal(cfg)

	tests := []struct {
		name       string
		busFunc    busFunc
		watchEvent store.WatchEvent
	}{
		{
			name:       "watch error",
			watchEvent: store.WatchEvent{Type: store.WatchError},
			busFunc: func(bus *mockbus.MockBus) {
				bus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
			},
		},
		{
			name: "invalid proto message",
			watchEvent: store.WatchEvent{
				Type:   store.WatchCreate,
				Object: []byte("foo"),
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
			},
		},
		{
			name: "invalid struct",
			watchEvent: store.WatchEvent{
				Type:   store.WatchCreate,
				Object: stateBytes,
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything)
			},
		},
		{
			name: "bus error",
			watchEvent: store.WatchEvent{
				Type:   store.WatchCreate,
				Object: cfgBytes,
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", mock.Anything, mock.Anything).Once().Return(errors.New("error"))
			},
		},
		{
			name: "watch events are successfully published to the bus",
			watchEvent: store.WatchEvent{
				Type:   store.WatchCreate,
				Object: cfgBytes,
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock a watcher
			watcher := &mockWatcher{resultChan: make(chan store.WatchEvent)}
			defer watcher.Stop()

			// Mock the bus
			bus := &mockbus.MockBus{}
			if tt.busFunc != nil {
				tt.busFunc(bus)
			}

			go handleResults(watcher, bus)

			watcher.resultChan <- tt.watchEvent
		})
	}
}
