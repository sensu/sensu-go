package monitor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	monitorPathPrefix     = "monitors"
	monitorKeyBuilder     = store.NewKeyBuilder(monitorPathPrefix)
	leasesToRevoke        = make(map[clientv3.LeaseID]struct{})
	leasesMu              sync.Mutex
	leasesOnce            sync.Once
	packageClientDoNotUse *clientv3.Client
)

func init() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// TODO:(echlebek) get rid of this when we fix all of this stuff :)
	go func() {
		<-sigs

		leasesMu.Lock()
		defer leasesMu.Unlock()

		for id := range leasesToRevoke {
			// There is nothing useful we can do with the error.
			// If the lease can't be revoked, it will expire later on.
			_, _ = packageClientDoNotUse.Lease.Revoke(context.TODO(), id)
		}
	}()
}

// Supervisor provides a way to refresh a named monitor. It is a proxy for all
// of the running monitors in the system.
type Supervisor interface {
	// Monitor starts a new monitor or resets an existing monitor.
	Monitor(ctx context.Context, id string, event *types.Event, ttl int64) error
}

// EtcdSupervisor is an etcd backend monitor supervisor based on leased keys.
// Each key has a watcher that waits for a DELETE or PUT event and calls a
// handler.
type EtcdSupervisor struct {
	failureHandler FailureHandler
	errorHandler   ErrorHandler
	client         *clientv3.Client
}

type monitor struct {
	key     string
	leaseID clientv3.LeaseID
	ttl     int64
}

// EtcdFactory returns a Factory bound to an etcd client
func EtcdFactory(c *clientv3.Client) Factory {
	return func(h Handler) Supervisor {
		return NewEtcdSupervisor(c, h)
	}
}

// Factory is a function that receives handlers and returns a Supervisor.
type Factory func(Handler) Supervisor

// NewEtcdSupervisor returns a new Supervisor backed by Etcd.
func NewEtcdSupervisor(client *clientv3.Client, h Handler) *EtcdSupervisor {
	leasesOnce.Do(func() {
		packageClientDoNotUse = client
	})
	return &EtcdSupervisor{
		client:         client,
		failureHandler: h,
		errorHandler:   h,
	}
}

// Monitor checks for the presence of a monitor for a given name.
// If no monitor exists, one is created. If a monitor exists, its lease ttl is
// extended. If the monitor's ttl has changed, a new lease is created and the
// key is updated with that new lease.
func (m *EtcdSupervisor) Monitor(ctx context.Context, name string, event *types.Event, ttl int64) error {
	if ttl == 0 {
		return errors.New("monitor ttl cannot be 0")
	}
	key := monitorKeyBuilder.Build(name)
	// try to get the monitor from the store
	mon, err := m.getMonitor(ctx, key)
	if err != nil {
		return err
	}
	// if it exists and the ttl matches the original ttl of the lease, extend its
	// lease with keep-alive.
	if mon != nil && mon.ttl == ttl {
		logger.Debugf("a lease for the key %s already exists, extending lease %v", key, mon.leaseID)
		_, kaerr := m.client.KeepAliveOnce(ctx, mon.leaseID)
		if kaerr == nil {
			return kaerr
		}
		logger.WithError(kaerr).Errorf("unable to extend lease %v, creating new lease for the key %s", mon.leaseID, key)
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
		err := m.failureHandler.HandleFailure(event)
		if err != nil {
			m.errorHandler.HandleError(err)
		}
	}

	shutdownFunc := func() {
		logger.Infof("shutting down monitor for %s", key)
	}

	// start the watcher
	watchMon(ctx, m.client, mon, failureFunc, shutdownFunc)
	logger.Debugf("starting a monitor for the key %s", key)

	return nil
}

func (m *EtcdSupervisor) getMonitor(ctx context.Context, key string) (*monitor, error) {
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
func watchMon(ctx context.Context, cli *clientv3.Client, mon *monitor, failureHandler func(), shutdownHandler func()) {
	if mon.ttl == 0 {
		panic("zero-duration ttl")
	}
	ctx, cancel := context.WithCancel(ctx)
	responseChan := cli.Watch(ctx, mon.key)
	timer := time.NewTimer(time.Duration(mon.ttl) * time.Second)
	go func() {
		defer timer.Stop()
		defer cancel()
		for {
			select {
			case wresp := <-responseChan:
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
			case <-timer.C:
				failureHandler()
				return
			}
		}
	}()
	leasesMu.Lock()
	leasesToRevoke[mon.leaseID] = struct{}{}
	leasesMu.Unlock()
}
