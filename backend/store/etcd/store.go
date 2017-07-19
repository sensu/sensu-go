package etcd

import (
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	etcdRoot = "/sensu.io"
)

// Store is an implementation of the sensu-go/backend/store.Store iface.
type etcdStore struct {
	client *clientv3.Client
	kvc    clientv3.KV
	etcd   *Etcd

	keepalivesPath string
}

// NewStore ...
func (e *Etcd) NewStore() (store.Store, error) {
	c, err := e.NewClient()
	if err != nil {
		return nil, err
	}

	store := &etcdStore{
		etcd:   e,
		client: c,
		kvc:    clientv3.NewKV(c),
	}

	store.keepalivesPath = path.Join(etcdRoot, keepalivesPathPrefix, store.etcd.cfg.Name)
	return store, nil
}
