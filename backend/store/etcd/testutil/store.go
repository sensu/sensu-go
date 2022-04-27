package testutil

import (
	"io/ioutil"
	"os"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// IntegrationTestStore wrapper for etcd & store
type IntegrationTestStore struct {
	*etcdstore.Store
	Client *clientv3.Client
	// underscores to avoid collision w/ store
	_etcd        *etcd.Etcd
	_removeTmpFn func()
}

// Teardown etcd and remove temp directory
func (e *IntegrationTestStore) Teardown() {
	_ = e._etcd.Shutdown()
	e._removeTmpFn()
}

// GetStore return etcd client
func (e *IntegrationTestStore) GetStore() store.Store {
	return e.Store
}

// NewStoreInstance returns new isolated store
func NewStoreInstance() (*IntegrationTestStore, error) {
	// Create temp dir
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		return nil, err
	}
	removeTmp := func() { _ = os.RemoveAll(tmpDir) }

	cfg := etcd.NewConfig()
	cfg.Name = "default"
	cfg.DataDir = tmpDir
	cfg.InitialClusterState = etcd.ClusterStateNew

	cfg.ListenClientURLs = []string{"http://127.0.0.1:0"}
	cfg.ListenPeerURLs = []string{"http://127.0.0.1:0"}
	cfg.InitialCluster = "default=http://127.0.0.1:0"
	cfg.AdvertiseClientURLs = cfg.ListenClientURLs
	cfg.InitialAdvertisePeerURLs = cfg.ListenPeerURLs
	cfg.LogLevel = "info"

	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		removeTmp()
		return nil, err
	}

	client := e.NewEmbeddedClient()

	st := etcdstore.NewStore(client)

	return &IntegrationTestStore{
		Store:        st,
		Client:       client,
		_etcd:        e,
		_removeTmpFn: removeTmp,
	}, nil
}

// RunWithStore starts new isolated etcd store, defers teardown and then runs
// given closure with store.
//
//  Ex.
//
//    RunWithStore(func (store store.Store) {
//      err := store.CreateCheck(...)
//      assert.NoError(err)
//    })
//
func RunWithStore(fn func(store.Store)) error {
	store, err := NewStoreInstance()
	if err != nil {
		return err
	}
	defer store.Teardown()

	fn(store.Store)
	return nil
}
