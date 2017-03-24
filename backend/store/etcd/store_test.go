package etcd

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEtcdStore(t *testing.T) {
	util.WithTempDir(func(tmpDir string) {
		ports := make([]int, 2)
		err := util.RandomPorts(ports)
		if err != nil {
			log.Panic(err)
		}
		clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
		apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
		initCluster := fmt.Sprintf("default=%s", apURL)
		fmt.Println(initCluster)

		cfg := NewConfig()
		cfg.StateDir = tmpDir
		cfg.ClientListenURL = clURL
		cfg.PeerListenURL = apURL
		cfg.InitialCluster = initCluster

		e, err := NewEtcd(cfg)
		if e != nil {
			defer e.Shutdown()
		}
		assert.NoError(t, err)
		if err != nil {
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			pprof.Lookup("threadcreate").WriteTo(os.Stdout, 1)
			pprof.Lookup("heap").WriteTo(os.Stdout, 1)
			assert.FailNow(t, "unable to start new etcd")
		}

		store, err := e.NewStore()
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t, "failed to get store from etcd")
		}

		entity := &types.Entity{
			ID: "0",
		}
		err = store.UpdateEntity(entity)
		assert.NoError(t, err)
		retrieved, err := store.GetEntityByID(entity.ID)
		assert.NoError(t, err)
		assert.Equal(t, entity.ID, retrieved.ID)
		entities, err := store.GetEntities()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(entities))
		err = store.DeleteEntity(entity)
		assert.NoError(t, err)
		retrieved, err = store.GetEntityByID(entity.ID)
		assert.Nil(t, retrieved)
		assert.NoError(t, err)
	})
}
