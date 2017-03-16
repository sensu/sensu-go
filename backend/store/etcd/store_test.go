package etcd

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEtcdStore(t *testing.T) {
	testWithTempDir(func(tmpDir string) {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Panic(err)
		}
		addr, err := net.ResolveTCPAddr("tcp", l.Addr().String())
		if err != nil {
			log.Panic(err)
		}
		clURL := fmt.Sprintf("http://127.0.0.1:%d", addr.Port)
		apURL := fmt.Sprintf("http://127.0.0.1:%d", addr.Port+1)
		initCluster := fmt.Sprintf("default=%s", apURL)
		fmt.Println(initCluster)
		l.Close()

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
		err = store.DeleteEntity(entity)
		assert.NoError(t, err)
		retrieved, err = store.GetEntityByID(entity.ID)
		assert.Nil(t, retrieved)
		assert.Error(t, err)
	})
}
