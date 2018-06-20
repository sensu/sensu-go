package monitor

import (
	"context"
	"fmt"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	monitorPathPrefix = "monitors"
	monitorKeyBuilder = store.NewKeyBuilder(monitorPathPrefix)
)

// Service is the monitors interface.
type Service interface {
	// RefreshMonitor starts a new monitor.
	RefreshMonitor(ctx context.Context, name string, entity *types.Entity, event *types.Event, ttl int64) error
}

// Factory takes an entity and returns a Monitor interface so the
// monitor can be mocked.
type Factory func(*clientv3.Client, UpdateHandler, FailureHandler) Service

// MonitorFailureHandler provides a failure handler. TODO: rename this to
// FailureHandler when we remove the other monitor code.
type MonitorFailureHandler interface {
	HandleFailure(entity *types.Entity, event *types.Event) error
}

type ErrorHandler interface {
	HandleError(error)
}

// ErrorHandler provides a handler for errors from WatchMon.
type ErrorHandlerFunc func(error)

func (e ErrorHandlerFunc) HandleError(err error) {
	e(err)
}

// EtcdService is an etcd backend monitor service based on leased keys. Each key
// has a watcher that waits for a DELETE or PUT event and calls a handler.
type EtcdService struct {
	failureHandler MonitorFailureHandler
	errorHandler   ErrorHandler
	client         *clientv3.Client
}

type monitor struct {
	key     string
	leaseID clientv3.LeaseID
	ttl     int64
}

// NewService returns a new monitor service.
func NewService(client *clientv3.Client, failureHandler MonitorFailureHandler, errorHandler ErrorHandler) *EtcdService {
	return &EtcdService{
		client:         client,
		failureHandler: failureHandler,
		errorHandler:   errorHandler,
	}
}

// RefreshMonitor checks for the presense of a monitor for a given name.
// If no monitor exists, one is created. If a monitor exists, its lease ttl is
// extended. If the monitor's ttl has changed, a new lease is created and the
// key is updated with that new lease.
func (m *EtcdService) RefreshMonitor(ctx context.Context, name string, entity *types.Entity, event *types.Event, ttl int64) error {
	key := monitorKeyBuilder.Build(name)
	// try to get the monitor from the store
	mon, err := m.getMonitor(ctx, key)
	if err != nil {
		return err
	}
	// if it exists and the ttl matches the original ttl of the lease, extend its
	// lease with keep-alive.
	if mon != nil && mon.ttl == ttl {
		_, kaerr := m.client.KeepAliveOnce(ctx, mon.leaseID)
		if kaerr != nil {
			return kaerr
		}

		return nil
	}

	// If the ttls do not match or the monitor doesn't exist, create a new lease
	// and do a put on the key with that lease.
	lease, err := m.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	mon = &monitor{
		key:     key,
		leaseID: lease.ID,
		ttl:     ttl,
	}

	_, err = m.client.Put(ctx, key, fmt.Sprintf("%d", mon.ttl), clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	failureFunc := func() {
		logger.Infof("monitor timed out, for %s, handling failure", key)
		err := m.failureHandler.HandleFailure(entity, event)
		if err != nil {
			m.errorHandler.HandleError(err)
		}
	}

	shutdownFunc := func() {
		logger.Info("shutting down monitor for %s", key)
	}

	// start the watcher
	watchMon(ctx, m.client, mon.key, failureFunc, shutdownFunc)
	return nil
}

func (m *EtcdService) getMonitor(ctx context.Context, key string) (*monitor, error) {
	// try to get the key from the store
	response, err := m.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	// if it exists, return it as a monitor
	if len(response.Kvs) > 0 {
		kv := response.Kvs[0]
		leaseID := clientv3.LeaseID(kv.Lease)
		ttl, err := strconv.ParseInt(string(kv.Value), 10, 64)
		if err != nil {
			return nil, err
		}
		return &monitor{
			key:     string(kv.Key),
			leaseID: leaseID,
			ttl:     ttl,
		}, nil
	}
	return nil, nil
}

// watchMon takes a monitor key and watches for etcd ops. If a DELETE event
// is witnessed, it calls the provided HandleFailure func. If a PUT event is
// witnessed, the watcher is stopped.
func watchMon(ctx context.Context, cli *clientv3.Client, key string, failureHandler func(), shutdownHandler func()) {
	go func() {
		responseChan := cli.Watch(ctx, key)
		for wresp := range responseChan {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					failureHandler()
					return
				}
				// if there is a PUT on the key, the lease has been extended,
				// and we want to kill this watcher to avoid duplicate watchers.
				if ev.Type == mvccpb.PUT {
					shutdownHandler()
					return
				}
			}
		}
	}()
}
