// +build integration,!race

package monitor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/coreos/etcd/clientv3"
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

func putKeyWithLease(cli *clientv3.Client, key string, ttl int64) error {
	lease, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), key, fmt.Sprintf("%d", ttl), clientv3.WithLease(lease.ID))
	return err
}

func TestGetMonitorNew(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testGetMonitorNew"
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewService(client, handler, handler)
	err = monitorService.GetMonitor(context.Background(), monitorName, testEntity, testEvent, 15)
	require.NoError(t, err)

}

func TestGetMonitorExisting(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testGetMonitorExisting"
	monitorPath := monitorKeyBuilder.Build(monitorName)
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewService(client, handler, handler)

	err = putKeyWithLease(client, monitorPath, 15)
	require.NoError(t, err)

	err = monitorService.GetMonitor(context.Background(), monitorName, testEntity, testEvent, 15)
	require.NoError(t, err)
}

func TestGetMonitorNewTTL(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testGetMonitorNewTTL"
	monitorPath := monitorKeyBuilder.Build(monitorName)
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewService(client, handler, handler)

	err = putKeyWithLease(client, monitorPath, 15)
	require.NoError(t, err)
	response, err := client.Get(context.Background(), monitorPath)
	require.NoError(t, err)
	fmt.Println("put response in test:", response)

	err = monitorService.GetMonitor(context.Background(), monitorName, testEntity, testEvent, 20)
	require.NoError(t, err)
}

func TestLittleGetMonitorNone(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	handler := &testMonitorsHandler{}
	monitorService := NewService(client, handler, handler)
	mon, err := monitorService.getMonitor(context.Background(), "testLittleGetMonitorNone")
	require.NoError(t, err)
	assert.Nil(t, mon)
}

func TestLittleGetMonitorExisting(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	handler := &testMonitorsHandler{}
	testMon := &monitor{
		key:     "testLittleGetMonitorExisting",
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
	require.NoError(t, err)
	defer client.Close()

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
	require.NoError(t, err)
	defer client.Close()

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
