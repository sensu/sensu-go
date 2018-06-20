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

// create leased key under monitors prefix:
//	- /sensu.io/monitors/org/env/<entity or round robin check>
// update leased key
// get ttl of current monitor (check for existence of key)
// watch monitor for deletion event, alert on that entity
// also might need a mechanism for deleting a monitor (ie deregistration)

var (
	monitorPathPrefix = "monitors"
	monitorKeyBuilder = store.NewKeyBuilder(monitorPathPrefix)
)

// EtcdGetter provides access to the etcd client.
type EtcdGetter struct {
	Client *clientv3.Client
}

// Service is the monitors interface.
type Service interface {
	// GetMonitor starts a new monitor.
	GetMonitor(ctx context.Context, name string, entity *types.Entity, event *types.Event, ttl int64) error
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

// EtcdService is an etcd backed monitor service monitor based on a leased key.
// Monitor leases can be extended, and a watcher on the key will run a handler
// when the lease expires and the key is deleted.
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

// NewMonitor returns a new monitor.
func NewService(client *clientv3.Client, failureHandler MonitorFailureHandler, errorHandler ErrorHandler) *EtcdService {
	return &EtcdService{
		client:         client,
		failureHandler: failureHandler,
		errorHandler:   errorHandler,
	}
}

// GetMonitor checks for the presense of a monitor for a given entity or check.
// if no monitor exists, one is created. If a monitor exists, its ttl is
// extended. If the monitor's ttl has changed, create a new lease.
func (m *EtcdService) GetMonitor(ctx context.Context, name string, entity *types.Entity, event *types.Event, ttl int64) error {
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

	// put key and start the watcher
	res, err := m.client.Put(ctx, key, fmt.Sprintf("%d", mon.ttl), clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	fmt.Println("transaction response in monitors:", res)

	failureFunc := func() {
		logger.Infof("monitor timed out, for %s, handling failure", key)
		err := m.failureHandler.HandleFailure(entity, event)
		if err != nil {
			fmt.Println("monitor timed out:", err)
			m.errorHandler.HandleError(err)
		}
	}

	shutdownFunc := func() {
		logger.Info("shutting down monitor for %s", key)
	}

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
		// if there is no lease, this will be 0 (NoLease constant in etcd)
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
	// otherwise, return nil
	return nil, nil
}

// watchMon takes a monitor key and watches for a deletion op. If a delete event
// is witnessed, it calls the provided HandleFailure func.
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
				// and we want to kill this watcher as another one will be
				// started.
				if ev.Type == mvccpb.PUT {
					shutdownHandler()
					return
				}
			}
		}
	}()
}
