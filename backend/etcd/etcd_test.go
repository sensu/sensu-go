// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEtcd(t *testing.T) {
	e, cleanup := NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	assert.NoError(t, err)
	kv := client.KV
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

	require.NoError(t, e.Shutdown())
}

func TestEtcdHealthy(t *testing.T) {
	e, cleanup := NewTestEtcd(t)
	defer cleanup()
	health := e.Healthy()
	assert.True(t, health)
}
