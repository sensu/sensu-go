package monitor

import (
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// create leased key under monitors prefix:
//	- /sensu.io/monitors/org/env/<entity or round robin check>
// update leased key
// get ttl of current monitor (check for existence of key)
// watch monitor for deletion event, alert on that entity

var (
	monitorPathPrefix = "monitors"
	monitorKeyBuilder = store.NewKeyBuilder(monitorPathPrefix)
)

// EtcdGetter is an Etcd implementation of Getter.
type EtcdGetter struct {
	*clientv3.Client
}

// Monitors is a monitor based on a leased key in etcd. Monitor leases can be
// extended, and a watcher on the key will run a handler when the lease expires
// and the key is deleted.
type Monitors struct {
	client *clientv3.Client
	mu     *sync.Mutex
}

type monitor struct {
	check  *types.Check
	entity *types.Entity
	event  *types.Event
}

// NewMonitors return new monitors.
func NewMonitors(client *clientv3.Client) *Monitors {
	return &Monitors{
		client: client,
	}
}

// GetMonitor returns a monitor. If a monitor is current, it is returned; if no
// monitor exists, it is created.
func (m *Monitors) GetMonitor(name string, ttl int) error {

	return nil
}

// UpdateMonitor extends the lease on a monitor.
func (m *Monitors) UpdateMonitor(name string, ttl int) error {
	return nil
}

// CheckMonitor checks for the existence of a monitor. It returns the ttl
// remaining if the monitor exists, otherwise it returns zero.
func (m *Monitors) CheckMonitor(name string) error {
	// might not need this func.
	return nil
}

// WatchMon takes a monitor key and watches for a deletion op. If a delete event
// is witnessed, it calls the provided HandleFailure func.
func (m *Monitors) WatchMon(name string) error {

	return nil
}
