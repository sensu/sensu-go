package monitor

import (
	"context"
	"fmt"

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

// MonitorUpdateHandler provides an event update handler.
type MonitorUpdateHandler interface {
	HandleUpdate(e *types.Event) error
}

// MonitorFailureHandler provides a failure handler.
type MonitorFailureHandler interface {
	HandleFailure(entity *types.Entity, event *types.Event) error
}

// Monitors is a monitor based on a leased key in etcd. Monitor leases can be
// extended, and a watcher on the key will run a handler when the lease expires
// and the key is deleted.
type Monitors struct {
	FailureHandler MonitorFailureHandler
	UpdateHandler  MonitorUpdateHandler
	client         *clientv3.Client
}

type monitor struct {
	entity  *types.Entity
	event   *types.Event
	key     string
	leaseID clientv3.LeaseID
	ttl     int64
	value   string
}

// NewMonitor returns a new monitor.
func NewMonitor(client *clientv3.Client) *Monitors {
	return &Monitors{
		client: client,
	}
}

// WatchMon takes a monitor key and watches for a deletion op. If a delete event
// is witnessed, it calls the provided HandleFailure func.
func (m *Monitors) WatchMon(ctx context.Context, mon *monitor) error {
	var err error
	// call a goroutine, on delete call HandleFailure
	go func() {
		responseChan := m.client.Watch(ctx, mon.key)
		for wresp := range responseChan {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					err = m.HandleFailure(mon.entity, mon.event)
					return
				}

			}
		}
	}()
	return err
}

// HandleFailure...
func (m *Monitors) HandleFailure(entity *types.Entity, event *types.Event) error {
	err := m.FailureHandler.HandleFailure(entity, event)
	return err
}

// GetMonitor checks for the presense of a monitor for a given entity or check.
// if no monitor exists, one is created. If a monitor exists, its ttl is
// extended. If the monitor's ttl has changed, it extended with a new lease.
func (m *Monitors) GetMonitor(ctx context.Context, name string, entity *types.Entity, event *types.Event, ttl int64) error {
	key := monitorKeyBuilder.WithContext(ctx).Build(name)
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
	leaseID := clientv3.LeaseID(ttl)

	lease, err := m.client.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	//create a monitor value
	value := "type of event or entity or check etc that we're storing"

	mon = &monitor{
		entity:  entity,
		event:   event,
		key:     key,
		leaseID: leaseID,
		ttl:     ttl,
		value:   value,
	}
	// put key to etcd and start the watcher - if a watcher already exists,
	// don't start it (figure out how to cancel a goroutine that's already
	// running? or maybe do this another way)
	cmp := clientv3.Compare(clientv3.Version(key), "=", 0)
	req := clientv3.OpPut(mon.key, mon.value, clientv3.WithLease(mon.leaseID))
	res, err := m.client.Txn(ctx).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}

	if !res.Succeeded {
		return fmt.Errorf("could not create monitor for %s", key)
	}
	// start watcher here
	err = m.WatchMon(ctx, mon)
	return err
}

func (m *Monitors) getMonitor(ctx context.Context, id string) (*monitor, error) {
	// try to get the key from the store
	response, err := m.client.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	// if it exists, return it as a monitor
	if len(response.Kvs) > 0 {
		mon := response.Kvs[0]
		leaseID := clientv3.LeaseID(mon.Lease)
		return &monitor{
			key:     string(mon.Key),
			leaseID: leaseID,
		}, nil
	}
	// otherwise, return nil
	return nil, nil
}
