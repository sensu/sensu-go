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
func (*testMonitorsHandler) HandleFailure(entity *types.Entity, event *types.Event) error {
	if entity.ID == "entity" {
		return nil
	}
	return errors.New("test failure handler error")
}

func (*testMonitorsHandler) HandleError(err error) {}

func putKeyWithLease(cli *clientv3.Client, key string, ttl int64) error {
	lease, err := cli.Grant(context.Background(), ttl)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), key, fmt.Sprintf("%d", ttl), clientv3.WithLease(lease.ID))
	return err
}

// TestRefreshMonitorNew
func TestRefreshMonitorNew(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testRefreshMonitorNew"
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewEtcdService(client, handler, handler)
	err = monitorService.RefreshMonitor(context.Background(), monitorName, testEntity, testEvent, 15)
	require.NoError(t, err)

}

func TestRefreshMonitorExisting(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testRefreshMonitorExisting"
	monitorPath := monitorKeyBuilder.Build(monitorName)
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewEtcdService(client, handler, handler)

	err = putKeyWithLease(client, monitorPath, 15)
	require.NoError(t, err)

	err = monitorService.RefreshMonitor(context.Background(), monitorName, testEntity, testEvent, 15)
	require.NoError(t, err)
}

func TestRefreshMonitorNewTTL(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testRefreshMonitorNewTTL"
	monitorPath := monitorKeyBuilder.Build(monitorName)
	testEntity := types.FixtureEntity("entity")
	testEvent := types.FixtureEvent(testEntity.ID, "testCheck")

	handler := &testMonitorsHandler{}
	monitorService := NewEtcdService(client, handler, handler)

	err = putKeyWithLease(client, monitorPath, 15)
	require.NoError(t, err)
	_, err = client.Get(context.Background(), monitorPath)
	require.NoError(t, err)

	err = monitorService.RefreshMonitor(context.Background(), monitorName, testEntity, testEvent, 20)
	require.NoError(t, err)
}

func TestGetMonitorNone(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	handler := &testMonitorsHandler{}
	monitorService := NewEtcdService(client, handler, handler)
	mon, err := monitorService.getMonitor(context.Background(), "testGetMonitorNone")
	require.NoError(t, err)
	assert.Nil(t, mon)
}

func TestGetMonitorExisting(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	handler := &testMonitorsHandler{}
	testMon := &monitor{
		key:     "testGetMonitorExisting",
		leaseID: 0,
		ttl:     0,
	}
	monitorService := NewEtcdService(client, handler, handler)
	_, err = client.Put(context.Background(), testMon.key, fmt.Sprintf("%d", testMon.ttl))
	require.NoError(t, err)

	mon, err := monitorService.getMonitor(context.Background(), testMon.key)
	require.NoError(t, err)
	assert.EqualValues(t, testMon.key, mon.key)
}

// TestWatchMonDelete uses a wait group to monitor the state of watchMon. The
// test passes if the failure handler is called, which closes the wait group.
func TestWatchMonDelete(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	failWait := &sync.WaitGroup{}

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

// TestWatchMonPut uses a wait group to monitor the state of watchMon. The
// test passes if the failure handler is called, which closes the wait group.
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
