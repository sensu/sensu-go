package monitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testMonitorsHandler struct{}

// create failure and error handlers for use with the monitor
func (handler *testMonitorsHandler) HandleFailure(entity *types.Entity, event *types.Event) error {
	if entity.ID == "entity" {
		return nil
	}
	return errors.New("test failure handler error")
}

func (handler *testMonitorsHandler) HandleError(err error) {
	fmt.Println("error handled")
}

func TestMonitorsHandleUpdate(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)
	passEntity := types.FixtureEntity("entity")
	failEntity := types.FixtureEntity("fail")

	testCases := []struct {
		name        string
		entity      *types.Entity
		event       *types.Event
		handler     *testMonitorsHandler
		expectedErr error
	}{
		{
			name:   "test no error",
			entity: passEntity,
			event: &types.Event{
				Entity: passEntity,
			},
			handler:     &testMonitorsHandler{},
			expectedErr: nil,
		},
		{
			name:   "test with error",
			entity: failEntity,
			event: &types.Event{
				Entity: failEntity,
			},
			handler:     &testMonitorsHandler{},
			expectedErr: errors.New("test failure handler error"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			monitorService := NewService(client, tc.handler, tc.handler)

			assert.EqualValues(tc.expectedErr, monitorService.failureHandler.HandleFailure(tc.entity, tc.event))
		})
	}

}

func TestWatchMonDelete(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	var failWait sync.WaitGroup

	testFailureHandler := func() {
		failWait.Done()
	}

	key := "monitortest"
	_, err = client.Put(context.Background(), key, "test value")
	require.NoError(t, err)
	watchMon(context.Background(), client, key, testFailureHandler, nil)
	failWait.Add(1)
	_, err = client.Delete(context.Background(), key)
	require.NoError(t, err)
	failWait.Wait()
}

func TestWatchMonPut(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	var shutdownWait sync.WaitGroup

	testShutdownHandler := func() {
		shutdownWait.Done()
	}

	key := "monitortest"
	_, err = client.Put(context.Background(), key, "test value")
	require.NoError(t, err)
	watchMon(context.Background(), client, key, nil, testShutdownHandler)
	shutdownWait.Add(1)
	_, err = client.Put(context.Background(), key, "test value")
	require.NoError(t, err)
	shutdownWait.Wait()
}
