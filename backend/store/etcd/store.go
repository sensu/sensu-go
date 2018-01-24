package etcd

import (
	"path"

	"github.com/coreos/etcd/clientv3"
)

const (
	// EtcdRoot is the root of all sensu storage.
	EtcdRoot = "/sensu.io"
)

// Store is an implementation of the sensu-go/backend/store.Store iface.
type Store struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *Etcd

	keepalivesPath string
}

// NewStore creates a new Store.
func (e *Etcd) NewStore() (*Store, error) {
	c, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	store := &Store{
		etcd:   e,
		client: c,
		kvc:    clientv3.NewKV(c),
	}

	store.keepalivesPath = path.Join(EtcdRoot, keepalivesPathPrefix, store.etcd.cfg.Name)
	return store, nil
}
