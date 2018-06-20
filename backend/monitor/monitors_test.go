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

func (handler *testMonitorsHandler) HandleError(err error) {}

func TestGetMonitor(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	monitorName := "testGetMonitor"
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewService(client, handler, handler)
	err = monitorService.GetMonitor(context.Background(), monitorName, testEntity, testEvent, 15)
	require.NoError(t, err)
}

func TestLittleGetMonitorNoExistingMonitor(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	handler := &testMonitorsHandler{}
	monitorService := NewService(client, handler, handler)
	mon, err := monitorService.getMonitor(context.Background(), "testLittleGetKey")
	require.NoError(t, err)
	assert.EqualValues(t, nil, mon)
}

func TestLittleGetMonitorWithExistingMonitor(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	handler := &testMonitorsHandler{}
	testMon := &monitor{
		key:     "testKey",
		leaseID: 0,
		ttl:     0,
	}
	monitorService := NewService(client, handler, handler)
	_, err = client.Put(context.Background(), testMon.key, fmt.Sprintf("%d", testMon.ttl))
	require.NoError(t, err)

	mon, err := monitorService.getMonitor(context.Background(), testMon.key)
	require.NoError(t, err)
	assert.EqualValues(t, testMon.key, mon.key)
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

	key := "monitorTestDelete"
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

	key := "monitorTestPut"
	_, err = client.Put(context.Background(), key, "test value")
	require.NoError(t, err)
	watchMon(context.Background(), client, key, nil, testShutdownHandler)
	shutdownWait.Add(1)
	_, err = client.Put(context.Background(), key, "test value")
	require.NoError(t, err)
	shutdownWait.Wait()
}
