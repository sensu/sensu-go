package etcd

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/assert"
)

func TestNewEtcd(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer os.RemoveAll(tmpDir)

	cfg := NewConfig()
	cfg.StateDir = tmpDir

	err = NewEtcd(cfg)
	assert.NoError(t, err)

	client, err := NewClient()
	kv := clientv3.NewKV(client)
	assert.NotNil(t, kv)

	putsResp, err := kv.Put(context.Background(), "key", "value")
	assert.NoError(t, err)
	assert.NotNil(t, putsResp)

	if putsResp == nil {
		assert.FailNow(t, "got nil put response from etcd")
	}

	getResp, err := kv.Get(context.Background(), "key")
	assert.NoError(t, err)
	assert.NotNil(t, getResp)

	if getResp == nil {
		assert.FailNow(t, "got nil get response from etcd")
	}
	assert.Equal(t, 1, len(getResp.Kvs))
	assert.Equal(t, "key", string(getResp.Kvs[0].Key))
	assert.Equal(t, "value", string(getResp.Kvs[0].Value))

	Shutdown()
}
