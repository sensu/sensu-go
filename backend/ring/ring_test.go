// +build integration

package ring

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	ring := EtcdGetter{client}.GetRing("testadd")
	err = ring.Add(context.Background(), "foo")
	assert.NoError(t, err)
}

func TestRemove(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	ring := EtcdGetter{client}.GetRing("testremove")
	require.NoError(t, ring.Add(context.Background(), "foo"))
	require.NoError(t, ring.Remove(context.Background(), "foo"))
}

func TestNext(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	ring := EtcdGetter{client}.GetRing("testnext")

	items := []string{"foo", "bar", "baz"}
	for _, item := range items {
		require.NoError(t, ring.Add(context.Background(), item))
	}

	for i := 0; i < 9; i++ {
		item, err := ring.Next(context.Background())
		require.NoError(t, err)
		assert.Equal(t, items[i%3], item)
	}

	require.NoError(t, ring.Remove(context.Background(), "bar"))

	newItems := []string{"foo", "baz"}

	for i := 0; i < 10; i++ {
		item, err := ring.Next(context.Background())
		require.NoError(t, err)
		assert.Equal(t, newItems[i%2], item)
	}
}

func TestPeek(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)

	ring := EtcdGetter{client}.GetRing("testpeek")

	items := []string{"foo", "bar", "baz"}
	for _, item := range items {
		require.NoError(t, ring.Add(context.Background(), item))
	}

	item, err := ring.Peek(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "foo", item)
}
