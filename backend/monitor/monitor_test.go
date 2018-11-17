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
func (*testMonitorsHandler) HandleFailure(event *types.Event) error {
	if event.Entity.Name == "entity" {
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

// TestMonitorNew
func TestMonitorNew(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testMonitorNew"
	testEvent := types.FixtureEvent("entity", "testCheck")

	handler := &testMonitorsHandler{}
	monitorSupervisor := NewEtcdSupervisor(client, handler)
	err = monitorSupervisor.Monitor(context.Background(), monitorName, testEvent, 15)
	require.NoError(t, err)

}

func TestMonitorExisting(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testMonitorExisting"
	monitorPath := monitorKeyBuilder.Build(monitorName)
	testEvent := types.FixtureEvent("entity", "testCheck")

	handler := &testMonitorsHandler{}
	monitorSupervisor := NewEtcdSupervisor(client, handler)

	err = putKeyWithLease(client, monitorPath, 15)
	require.NoError(t, err)

	err = monitorSupervisor.Monitor(context.Background(), monitorName, testEvent, 15)
	require.NoError(t, err)
}

func TestMonitorNewTTL(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	monitorName := "testMonitorNewTTL"
	monitorPath := monitorKeyBuilder.Build(monitorName)
	testEvent := types.FixtureEvent("entity", "testCheck")

	handler := &testMonitorsHandler{}
	monitorSupervisor := NewEtcdSupervisor(client, handler)

	err = putKeyWithLease(client, monitorPath, 15)
	require.NoError(t, err)
	_, err = client.Get(context.Background(), monitorPath)
	require.NoError(t, err)

	err = monitorSupervisor.Monitor(context.Background(), monitorName, testEvent, 20)
	require.NoError(t, err)
}

func TestGetMonitorNone(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	handler := &testMonitorsHandler{}
	monitorSupervisor := NewEtcdSupervisor(client, handler)
	mon, err := monitorSupervisor.getMonitor(context.Background(), "testGetMonitorNone")
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
		ttl:     3600,
	}
	monitorSupervisor := NewEtcdSupervisor(client, handler)
	_, err = client.Put(context.Background(), testMon.key, fmt.Sprintf("%d", testMon.ttl))
	require.NoError(t, err)

	mon, err := monitorSupervisor.getMonitor(context.Background(), testMon.key)
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

	mon := &monitor{
		key:     "monitorTestDelete",
		leaseID: 0,
		ttl:     3600,
	}
	_, err = client.Put(context.Background(), mon.key, "test value")
	require.NoError(t, err)
	watchMon(context.Background(), client, mon, testFailureHandler, nil)
	failWait.Add(1)
	_, err = client.Delete(context.Background(), mon.key)
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

	mon := &monitor{
		key:     "monitorTestPut",
		leaseID: 0,
		ttl:     3600,
	}

	_, err = client.Put(context.Background(), mon.key, "test value")
	require.NoError(t, err)
	watchMon(context.Background(), client, mon, nil, testShutdownHandler)
	shutdownWait.Add(1)
	_, err = client.Put(context.Background(), mon.key, "test value")
	require.NoError(t, err)
	shutdownWait.Wait()
}
