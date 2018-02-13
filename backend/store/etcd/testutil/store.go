package testutil

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/testing/testutil"
)

// IntegrationTestStore wrapper for etcd & store
type IntegrationTestStore struct {
	*etcdstore.Store
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

	p := make([]int, 2)
	perr := testutil.RandomPorts(p)
	if perr != nil {
		removeTmp()
		return nil, perr
	}

	cfg := etcd.NewConfig()
	cfg.DataDir = tmpDir

	peerURL := fmt.Sprintf("http://127.0.0.1:%d", p[1])

	cfg.ListenClientURL = fmt.Sprintf("http://127.0.0.1:%d", p[0])
	cfg.ListenPeerURL = peerURL
	cfg.InitialCluster = fmt.Sprintf("default=http://127.0.0.1:%d", p[1])
	cfg.InitialAdvertisePeerURL = peerURL
	e, err := etcd.NewEtcd(cfg)
	if err != nil {
		removeTmp()
		return nil, err
	}

	st, err := etcdstore.NewStore(e)
	if err != nil {
		_ = e.Shutdown()
		removeTmp()
		return nil, err
	}

	return &IntegrationTestStore{
		Store:        st,
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
