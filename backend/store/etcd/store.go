package etcd

import (
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/etcd"
)

const (
	// EtcdRoot is the root of all sensu storage.
	EtcdRoot = "/sensu.io"
)

// Store is an implementation of the sensu-go/backend/store.Store iface.
type Store struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *etcd.Etcd

	keepalivesPath string
}

// NewStore creates a new Store.
func NewStore(e *etcd.Etcd) (*Store, error) {
	c, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	store := &Store{
		etcd:   e,
		client: c,
		kvc:    clientv3.NewKV(c),
	}

	store.keepalivesPath = path.Join(EtcdRoot, keepalivesPathPrefix, store.etcd.Name())
	return store, nil
}
